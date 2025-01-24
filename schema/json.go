package schema

import (
	"encoding/json"
)

// MarshalJSON return custom JSON byte
func (t Table) MarshalJSON() ([]byte, error) {
	referencedTables := []string{}
	for _, rt := range t.ReferencedTables {
		referencedTables = append(referencedTables, rt.Name)
	}

	return json.Marshal(&struct {
		Name             string        `json:"name"`
		Type             string        `json:"type"`
		Comment          string        `json:"comment,omitempty"`
		Columns          []*Column     `json:"columns"`
		Indexes          []*Index      `json:"indexes,omitempty"`
		Constraints      []*Constraint `json:"constraints,omitempty"`
		Triggers         []*Trigger    `json:"triggers,omitempty"`
		Def              string        `json:"def,omitempty"`
		Labels           Labels        `json:"labels,omitempty"`
		ReferencedTables []string      `json:"referenced_tables,omitempty"`
	}{
		Name:             t.Name,
		Type:             t.Type,
		Comment:          t.Comment,
		Columns:          t.Columns,
		Indexes:          t.Indexes,
		Constraints:      t.Constraints,
		Triggers:         t.Triggers,
		Def:              t.Def,
		Labels:           t.Labels,
		ReferencedTables: referencedTables,
	})
}

// MarshalJSON return custom JSON byte
func (c Column) MarshalJSON() ([]byte, error) {
	if c.Default.Valid {
		return json.Marshal(&struct {
			Name            string      `json:"name"`
			Type            string      `json:"type"`
			Nullable        bool        `json:"nullable"`
			Default         *string     `json:"default,omitempty"`
			ExtraDef        string      `json:"extra_def,omitempty"`
			Labels          Labels      `json:"labels,omitempty"`
			Comment         string      `json:"comment"`
			ParentRelations []*Relation `json:"-"`
			ChildRelations  []*Relation `json:"-"`
		}{
			Name:            c.Name,
			Type:            c.Type,
			Nullable:        c.Nullable,
			Default:         &c.Default.String,
			Comment:         c.Comment,
			ExtraDef:        c.ExtraDef,
			Labels:          c.Labels,
			ParentRelations: c.ParentRelations,
			ChildRelations:  c.ChildRelations,
		})
	}
	return json.Marshal(&struct {
		Name            string      `json:"name"`
		Type            string      `json:"type"`
		Nullable        bool        `json:"nullable"`
		Default         *string     `json:"default,omitempty"`
		Comment         string      `json:"comment,omitempty"`
		ExtraDef        string      `json:"extra_def,omitempty"`
		Labels          Labels      `json:"labels,omitempty"`
		ParentRelations []*Relation `json:"-"`
		ChildRelations  []*Relation `json:"-"`
	}{
		Name:            c.Name,
		Type:            c.Type,
		Nullable:        c.Nullable,
		Default:         nil,
		Comment:         c.Comment,
		ExtraDef:        c.ExtraDef,
		Labels:          c.Labels,
		ParentRelations: c.ParentRelations,
		ChildRelations:  c.ChildRelations,
	})
}

// MarshalJSON return custom JSON byte
func (r Relation) MarshalJSON() ([]byte, error) {
	columns := []string{}
	parentColumns := []string{}
	for _, c := range r.Columns {
		columns = append(columns, c.Name)
	}
	for _, c := range r.ParentColumns {
		parentColumns = append(parentColumns, c.Name)
	}

	return json.Marshal(&struct {
		Table             string   `json:"table"`
		Columns           []string `json:"columns"`
		Cardinality       string   `json:"cardinality,omitempty"`
		ParentTable       string   `json:"parent_table"`
		ParentColumns     []string `json:"parent_columns"`
		ParentCardinality string   `json:"parent_cardinality,omitempty"`
		Def               string   `json:"def"`
		Virtual           bool     `json:"virtual,omitempty"`
	}{
		Table:             r.Table.Name,
		Columns:           columns,
		Cardinality:       r.Cardinality.String(),
		ParentTable:       r.ParentTable.Name,
		ParentColumns:     parentColumns,
		ParentCardinality: r.ParentCardinality.String(),
		Def:               r.Def,
		Virtual:           r.Virtual,
	})
}

// UnmarshalJSON unmarshal JSON to schema.Table
func (t *Table) UnmarshalJSON(data []byte) error {
	s := struct {
		Name             string        `json:"name"`
		Type             string        `json:"type"`
		Comment          string        `json:"comment,omitempty"`
		Columns          []*Column     `json:"columns"`
		Indexes          []*Index      `json:"indexes,omitempty"`
		Constraints      []*Constraint `json:"constraints,omitempty"`
		Triggers         []*Trigger    `json:"triggers,omitempty"`
		Def              string        `json:"def,omitempty"`
		Labels           Labels        `json:"labels,omitempty"`
		ReferencedTables []string      `json:"referenced_tables,omitempty"`
	}{}
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	t.Name = s.Name
	t.Type = s.Type
	t.Comment = s.Comment
	t.Columns = s.Columns
	t.Indexes = s.Indexes
	t.Constraints = s.Constraints
	t.Triggers = s.Triggers
	t.Def = s.Def
	t.Labels = s.Labels
	for _, rt := range s.ReferencedTables {
		t.ReferencedTables = append(t.ReferencedTables, &Table{
			Name: rt,
		})
	}
	return nil
}

// UnmarshalJSON unmarshal JSON to schema.Column
func (c *Column) UnmarshalJSON(data []byte) error {
	s := struct {
		Name            string      `json:"name"`
		Type            string      `json:"type"`
		Nullable        bool        `json:"nullable"`
		Default         *string     `json:"default,omitempty"`
		Comment         string      `json:"comment,omitempty"`
		ExtraDef        string      `json:"extra_def,omitempty"`
		Labels          Labels      `json:"labels,omitempty"`
		ParentRelations []*Relation `json:"-"`
		ChildRelations  []*Relation `json:"-"`
	}{}
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	c.Name = s.Name
	c.Type = s.Type
	c.Nullable = s.Nullable
	if s.Default != nil {
		c.Default.Valid = true
		c.Default.String = *s.Default
	} else {
		c.Default.Valid = false
		c.Default.String = ""
	}
	c.ExtraDef = s.ExtraDef
	c.Labels = s.Labels
	c.Comment = s.Comment
	return nil
}

// UnmarshalJSON unmarshal JSON to schema.Relation
func (r *Relation) UnmarshalJSON(data []byte) error {
	s := struct {
		Table             string   `json:"table"`
		Columns           []string `json:"columns"`
		Cardinality       string   `json:"cardinality,omitempty"`
		ParentTable       string   `json:"parent_table"`
		ParentColumns     []string `json:"parent_columns"`
		ParentCardinality string   `json:"parent_cardinality,omitempty"`
		Def               string   `json:"def"`
		Virtual           bool     `json:"virtual,omitempty"`
	}{}
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	r.Table = &Table{
		Name: s.Table,
	}
	r.Columns = []*Column{}
	for _, c := range s.Columns {
		r.Columns = append(r.Columns, &Column{
			Name: c,
		})
	}
	r.Cardinality, err = ToCardinality(s.Cardinality)
	if err != nil {
		return err
	}
	r.ParentTable = &Table{
		Name: s.ParentTable,
	}
	r.ParentColumns = []*Column{}
	for _, c := range s.ParentColumns {
		r.ParentColumns = append(r.ParentColumns, &Column{
			Name: c,
		})
	}
	r.ParentCardinality, err = ToCardinality(s.ParentCardinality)
	if err != nil {
		return err
	}
	r.Def = s.Def
	r.Virtual = s.Virtual
	return nil
}
