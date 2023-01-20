package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/k1LoW/tbls/schema"
)

func TestLoadDefault(t *testing.T) {
	configFilepath := filepath.Join(testdataDir(), "empty.yml")
	config, err := New()
	if err != nil {
		t.Fatal(err)
	}
	err = config.Load(configFilepath)
	if err != nil {
		t.Fatal(err)
	}

	if want := ""; config.DSN.URL != want {
		t.Errorf("got %v\nwant %v", config.DSN.URL, want)
	}
	if want := "dbdoc"; config.DocPath != want {
		t.Errorf("got %v\nwant %v", config.DocPath, want)
	}
	if want := "svg"; config.ER.Format != want {
		t.Errorf("got %v\nwant %v", config.ER.Format, want)
	}
	if want := 1; *config.ER.Distance != want {
		t.Errorf("got %v\nwant %v", config.ER.Distance, want)
	}
}

func TestLoadConfigFile(t *testing.T) {
	t.Setenv("TBLS_TEST_PG_PASS", "pgpass")
	t.Setenv("TBLS_TEST_PG_DOC_PATH", "sample/pg")
	configFilepath := filepath.Join(testdataDir(), "config_test_tbls_2.yml")
	config, err := New()
	if err != nil {
		t.Fatal(err)
	}
	err = config.LoadConfigFile(configFilepath)
	if err != nil {
		t.Fatal(err)
	}

	if want := "pg://root:pgpass@localhost:55432/testdb?sslmode=disable"; config.DSN.URL != want {
		t.Errorf("got %v\nwant %v", config.DSN.URL, want)
	}

	if want := "sample/pg"; config.DocPath != want {
		t.Errorf("got %v\nwant %v", config.DocPath, want)
	}

	if want := "INDEX"; config.MergedDict.Lookup("Indexes") != want {
		t.Errorf("got %v\nwant %v", config.MergedDict.Lookup("Indexes"), want)
	}
}

func TestDuplicateConfigFile(t *testing.T) {
	config := &Config{
		root: filepath.Join(testdataDir(), "config"),
	}
	got := config.LoadConfigFile("")
	want := "duplicate config file [.tbls.yml, tbls.yml]"
	if fmt.Sprintf("%v", got) != want {
		t.Errorf("got %v\nwant %v", got, want)
	}
}

func TestMergeAditionalData(t *testing.T) {
	s := schema.Schema{
		Name: "testschema",
		Tables: []*schema.Table{
			&schema.Table{
				Name:    "users",
				Comment: "users comment",
				Columns: []*schema.Column{
					&schema.Column{
						Name: "id",
						Type: "serial",
					},
					&schema.Column{
						Name: "username",
						Type: "text",
					},
				},
				Indexes: []*schema.Index{
					&schema.Index{
						Name: "user_index",
					},
				},
				Constraints: []*schema.Constraint{
					&schema.Constraint{
						Name: "PRIMARY",
					},
				},
			},
			&schema.Table{
				Name:    "posts",
				Comment: "posts comment",
				Columns: []*schema.Column{
					&schema.Column{
						Name: "id",
						Type: "serial",
					},
					&schema.Column{
						Name: "user_id",
						Type: "int",
					},
					&schema.Column{
						Name: "title",
						Type: "text",
					},
				},
				Triggers: []*schema.Trigger{
					&schema.Trigger{
						Name: "update_posts_title",
					},
				},
			},
		},
	}
	c, err := New()
	if err != nil {
		t.Error(err)
	}
	err = c.LoadConfigFile(filepath.Join(testdataDir(), "config_test_tbls.yml"))
	if err != nil {
		t.Error(err)
	}
	err = c.MergeAdditionalData(&s)
	if err != nil {
		t.Error(err)
	}
	if want := 1; len(s.Relations) != want {
		t.Errorf("got %v\nwant %v", len(s.Relations), want)
	}
	users, _ := s.FindTableByName("users")
	posts, _ := s.FindTableByName("posts")
	title, _ := posts.FindColumnByName("title")
	if want := "post title"; title.Comment != want {
		t.Errorf("got %v\nwant %v", title.Comment, want)
	}

	index, err := users.FindIndexByName("user_index")
	if err != nil {
		t.Fatal(err)
	}
	if want := "user index"; index.Comment != want {
		t.Errorf("got %v want %v", index.Comment, want)
	}

	constraint, err := users.FindConstraintByName("PRIMARY")
	if err != nil {
		t.Fatal(err)
	}
	if want := "PRIMARY(id)"; constraint.Comment != want {
		t.Errorf("got %v want %v", constraint.Comment, want)
	}

	trigger, err := posts.FindTriggerByName("update_posts_title")
	if err != nil {
		t.Fatal(err)
	}
	if want := "update posts title"; trigger.Comment != want {
		t.Errorf("got %v want %v", trigger.Comment, want)
	}
}

func TestFilterTables(t *testing.T) {
	c, err := New()
	if err != nil {
		t.Error(err)
	}
	tests := []struct {
		include       []string
		exclude       []string
		labels        []string
		distance      int
		wantTables    int
		wantRelations int
	}{
		{[]string{}, []string{}, []string{}, 0, 5, 3},
		{[]string{}, []string{"schema_migrations"}, []string{}, 0, 4, 3},
		{[]string{}, []string{"users"}, []string{}, 0, 4, 1},
		{[]string{"users"}, []string{}, []string{}, 0, 1, 0},
		{[]string{"user*"}, []string{}, []string{}, 0, 2, 1},
		{[]string{"*options"}, []string{}, []string{}, 0, 1, 0},
		{[]string{"*"}, []string{"user_options"}, []string{}, 0, 4, 2},
		{[]string{"not_exist"}, []string{}, []string{}, 0, 0, 0},
		{[]string{"not_exist", "*"}, []string{}, []string{}, 0, 5, 3},
		{[]string{"users"}, []string{"*"}, []string{}, 0, 1, 0},
		{[]string{"use*"}, []string{"use*"}, []string{}, 0, 2, 1},
		{[]string{"use*"}, []string{"user*"}, []string{}, 0, 0, 0},
		{[]string{"user*"}, []string{"user_*"}, []string{}, 0, 1, 0},
		{[]string{"*", "user*"}, []string{"user_*"}, []string{}, 0, 4, 2},

		{[]string{"users"}, []string{}, []string{}, 1, 3, 2},
		{[]string{"user_options"}, []string{}, []string{}, 1, 2, 1},
		{[]string{"user_options"}, []string{}, []string{}, 2, 3, 2},
		{[]string{"user_options"}, []string{}, []string{}, 3, 4, 3},
		{[]string{}, []string{}, []string{}, 9, 5, 3},
		{[]string{"posts"}, []string{}, []string{}, 9, 4, 3},
		{[]string{""}, []string{"*"}, []string{}, 9, 0, 0},

		{[]string{}, []string{}, []string{"private"}, 0, 2, 1},
		{[]string{}, []string{}, []string{"option"}, 0, 2, 0},
		{[]string{}, []string{}, []string{"public", "private"}, 0, 4, 3},
		{[]string{}, []string{"users"}, []string{"private"}, 0, 1, 0},
		{[]string{}, []string{"user*"}, []string{"option"}, 0, 1, 0},
		{[]string{"users"}, []string{}, []string{"private"}, 0, 2, 1},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d.%v%v", i, tt.include, tt.exclude), func(t *testing.T) {
			s := newSchemaForTestFilterTables(t)
			c.Include = tt.include
			c.Exclude = tt.exclude
			c.includeLabels = tt.labels
			c.Distance = tt.distance
			err = c.FilterTables(s)
			if err != nil {
				t.Error(err)
			}
			if got := len(s.Tables); got != tt.wantTables {
				t.Errorf("got %v\nwant %v", got, tt.wantTables)
			}
			if got := len(s.Relations); got != tt.wantRelations {
				t.Errorf("got %v\nwant %v", got, tt.wantRelations)
			}
		})
	}
}

func TestModifySchema(t *testing.T) {
	s := schema.Schema{
		Name: "testschema",
		Tables: []*schema.Table{
			&schema.Table{
				Name:    "users",
				Comment: "users comment",
				Columns: []*schema.Column{
					&schema.Column{
						Name: "id",
						Type: "serial",
					},
					&schema.Column{
						Name: "username",
						Type: "text",
					},
				},
				Indexes: []*schema.Index{
					&schema.Index{
						Name: "user_index",
					},
				},
				Constraints: []*schema.Constraint{
					&schema.Constraint{
						Name: "PRIMARY",
					},
				},
			},
			&schema.Table{
				Name:    "posts",
				Comment: "posts comment",
				Columns: []*schema.Column{
					&schema.Column{
						Name: "id",
						Type: "serial",
					},
					&schema.Column{
						Name: "user_id",
						Type: "int",
					},
					&schema.Column{
						Name: "title",
						Type: "text",
					},
				},
				Triggers: []*schema.Trigger{
					&schema.Trigger{
						Name: "update_posts_title",
					},
				},
			},
			&schema.Table{
				Name: "migrations",
				Columns: []*schema.Column{
					&schema.Column{
						Name: "id",
						Type: "serial",
					},
					&schema.Column{
						Name: "name",
						Type: "text",
					},
				},
			},
		},
	}
	c, err := New()
	if err != nil {
		t.Error(err)
	}
	err = c.LoadConfigFile(filepath.Join(testdataDir(), "config_test_tbls.yml"))
	if err != nil {
		t.Error(err)
	}
	err = c.ModifySchema(&s)
	if err != nil {
		t.Error(err)
	}

	if want := 1; len(s.Relations) != want {
		t.Errorf("got %v\nwant %v", len(s.Relations), want)
	}
	posts, _ := s.FindTableByName("posts")
	title, _ := posts.FindColumnByName("title")
	if want := "post title"; title.Comment != want {
		t.Errorf("got %v\nwant %v", title.Comment, want)
	}
	if want := 2; len(title.Labels) != want {
		t.Errorf("got %v\nwant %v", len(title.Labels), want)
	}
	if want := 2; len(s.Tables) != want {
		t.Errorf("got %v\nwant %v", len(s.Tables), want)
	}
	if want := "mydatabase"; s.Name != want {
		t.Errorf("got %v\nwant %v", s.Name, want)
	}
}

func TestMaskedDSN(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{
			"pg://root:pgpass@localhost:5432/testdb?sslmode=disable",
			"pg://root:*****@localhost:5432/testdb?sslmode=disable",
		},
		{
			"pg://root@localhost:5432/testdb?sslmode=disable",
			"pg://root@localhost:5432/testdb?sslmode=disable",
		},
		{
			"pg://localhost:5432/testdb?sslmode=disable",
			"pg://localhost:5432/testdb?sslmode=disable",
		},
		{
			"bq://project-id/dataset-id?creds=/path/to/google_application_credentials.json",
			"bq://project-id/dataset-id?creds=/path/to/google_application_credentials.json",
		},
	}

	for _, tt := range tests {
		config, err := New()
		if err != nil {
			t.Fatal(err)
		}
		config.DSN.URL = tt.url
		got, err := config.MaskedDSN()
		if err != nil {
			t.Fatal(err)
		}
		if got != tt.want {
			t.Errorf("got %v\nwant %v", got, tt.want)
		}
	}
}

func testdataDir() string {
	wd, _ := os.Getwd()
	dir, _ := filepath.Abs(filepath.Join(filepath.Dir(wd), "testdata"))
	return dir
}

func Test_mergeDetectedRelations(t *testing.T) {
	var (
		err          error
		table        *schema.Table
		column       *schema.Column
		parentColumn *schema.Column
		relations    []*schema.Relation
	)
	s1 := &schema.Schema{
		Name: "testschema",
		Tables: []*schema.Table{
			{
				Name:    "users",
				Comment: "users comment",
				Columns: []*schema.Column{
					{
						Name: "id",
						Type: "serial",
					},
					{
						Name: "username",
						Type: "text",
					},
				},
			},
			{
				Name:    "posts",
				Comment: "posts comment",
				Columns: []*schema.Column{
					{
						Name: "id",
						Type: "serial",
					},
					{
						Name: "user_id",
						Type: "int",
					},
					{
						Name: "title",
						Type: "text",
					},
				},
			},
		},
	}
	s2 := &schema.Schema{
		Name: "testschema",
		Tables: []*schema.Table{
			{
				Name:    "users",
				Comment: "users comment",
				Columns: []*schema.Column{
					{
						Name: "id",
						Type: "serial",
					},
				},
			},
			{
				Name:    "posts",
				Comment: "posts comment",
				Columns: []*schema.Column{
					{
						Name: "id",
						Type: "serial",
					},
					{
						Name: "uid",
						Type: "int",
					},
					{
						Name: "title",
						Type: "text",
					},
				},
			},
		},
	}
	table, err = s1.FindTableByName("posts")
	if err != nil {
		t.Fatal(err)
	}
	column, err = table.FindColumnByName("user_id")
	if err != nil {
		t.Fatal(err)
	}

	relation := &schema.Relation{
		Virtual: true,
		Def:     "Detected Relation",
		Table:   table,
	}
	strategy, err := SelectNamingStrategy("default")
	if err != nil {
		t.Fatal(err)
	}
	if relation.ParentTable, err = s1.FindTableByName(strategy.ParentTableName("user_id")); err != nil {
		t.Fatal(err)
	}
	if parentColumn, err = relation.ParentTable.FindColumnByName(strategy.ParentColumnName("users")); err != nil {
		t.Fatal(err)
	}
	relation.Columns = append(relation.Columns, column)
	relation.ParentColumns = append(relation.ParentColumns, parentColumn)

	column.ParentRelations = append(column.ParentRelations, relation)
	parentColumn.ChildRelations = append(parentColumn.ChildRelations, relation)

	relations = append(relations, relation)

	type args struct {
		s *schema.Schema
	}
	type want struct {
		r []*schema.Relation
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Detect relation succeed",
			args: args{
				s: s1,
			},
			want: want{
				r: relations,
			},
		},
		{
			name: "Detect relation failed",
			args: args{
				s: s2,
			},
			want: want{
				r: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mergeDetectedRelations(tt.args.s, strategy)
			if !reflect.DeepEqual(tt.args.s.Relations, tt.want.r) {
				t.Errorf("got: %#v\nwant: %#v", tt.args.s.Relations, tt.want.r)
			}
		})
	}
}

func TestCheckVersion(t *testing.T) {
	tests := []struct {
		v    string
		c    string
		want error
	}{
		{"1.42.3", ">= 1.42", nil},
		{"1.42.3", ">= 1.42, < 2", nil},
		{"1.42.3", "> 1.42", nil},
		{"1.42.3", "1.42.3", nil},
		{"1.42.3", "1.42.4", errors.New("the required tbls version for the configuration is '1.42.4'. however, the running tbls version is '1.42.3'")},
	}
	for _, tt := range tests {
		cfg, err := New()
		if err != nil {
			t.Fatal(err)
		}
		cfg.RequiredVersion = tt.c
		if got := cfg.checkVersion(tt.v); fmt.Sprintf("%s", got) != fmt.Sprintf("%s", tt.want) {
			t.Errorf("got %v\nwant %v", got, tt.want)
		}
	}
}

func newSchemaForTestFilterTables(t *testing.T) *schema.Schema {
	t.Helper()
	s := &schema.Schema{}
	file, err := os.Open(filepath.Join(testdataDir(), "filter_tables.json"))
	if err != nil {
		t.Fatal(err)
	}
	dec := json.NewDecoder(file)
	if err := dec.Decode(s); err != nil {
		t.Fatal(err)
	}
	if err := s.Repair(); err != nil {
		t.Fatal(err)
	}
	return s
}
