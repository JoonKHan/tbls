# tbls [![Build Status](https://travis-ci.org/k1LoW/tbls.svg?branch=master)](https://travis-ci.org/k1LoW/tbls) [![GitHub release](https://img.shields.io/github/release/k1LoW/tbls.svg)](https://github.com/k1LoW/tbls/releases)


`tbls` is a tool for document a database, written in Go.

The key of Features of `tbls` are:

- Single binary
- Document in GitHub Friendly Markdown format
- CI friendly

[Usage](#usage) | [Sample](sample/) | [Integration with CI tools](#integration-with-ci-tools) | [Installation](#installation) | [Database Support](#database-support)

## Usage

```console
$ tbls
tbls is a tool for document a database, written in Go.

Usage:
  tbls [command]

Available Commands:
  diff        diff database and document
  doc         document a database
  help        Help about any command
  version     print tbls version

Flags:
  -h, --help   help for tbls

Use "tbls [command] --help" for more information about a command.
```

### Document a database schema

`tbls doc` analyzes a database and generate document in GitHub Friendly Markdown format.

```console
$ tbls doc postgres://user:pass@hostname:5432/dbname ./dbdoc
```

Sample [document](sample/) and [schema](test/pg.sql).

### Diff database schema and document

`tbls diff` shows the difference between database schema and generated document.

```console
$ tbls diff postgres://user:pass@hostname:5432/dbname ./dbdoc
```

## Integration with CI tools

1. Commit document using `tbls doc`.
2. Check document updates using `tbls diff`

Set following code to [`your-ci-config.yml`](.travis.yml).

```sh
DIFF=`tbls diff postgres://user:pass@localhost:5432/testdb?sslmode=disable ./sample` && if [ ! -z "$DIFF" ]; then echo "document does not match database." >&2 ; echo tbls diff postgres://user:pass@localhost:5432/testdb?sslmode=disable ./sample; exit 1; fi
```

Makefile sample is following

``` makefile
doc: ## Document database schema
	tbls doc postgres://user:pass@localhost:5432/testdb?sslmode=disable ./doc

testdoc: ## Test database schema document
	$(eval DIFF := $(shell tbls diff postgres://user:pass@localhost:5432/testdb?sslmode=disable ./doc))
	@test -z "$(DIFF)" || (echo "document does not match database." && postgres://user:pass@localhost:5432/testdb?sslmode=disable ./doc && exit 1)
```

**Tips:** If the order of the columns does not match, you can use the `--sort` option.

## Installation

```console
$ go get github.com/k1LoW/tbls
```

## Database Support

- PostgreSQL
