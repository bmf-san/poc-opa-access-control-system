# public.examples

## Description

## Columns

| Name | Type | Default | Nullable | Children | Parents | Comment |
| ---- | ---- | ------- | -------- | -------- | ------- | ------- |
| id | integer | nextval('examples_id_seq'::regclass) | false |  |  |  |
| name | varchar(100) |  | false |  |  |  |

## Constraints

| Name | Type | Definition |
| ---- | ---- | ---------- |
| examples_pkey | PRIMARY KEY | PRIMARY KEY (id) |

## Indexes

| Name | Definition |
| ---- | ---------- |
| examples_pkey | CREATE UNIQUE INDEX examples_pkey ON public.examples USING btree (id) |

## Relations

![er](public.examples.svg)

---

> Generated by [tbls](https://github.com/k1LoW/tbls)
