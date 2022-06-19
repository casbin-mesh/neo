# Neo Column Storage Brief Design

## Background

In order to prevent the massive memory cost when policy data grow, we decide to introduce column storage based on LSM to provide policy-disk-oriented management.

- Note that this document only focuses on the disk storage design, rather than the in-memory data structure.
- Currently no plan to support transactions, multiple index keys, or schema modification. Only supports several models’ policy CURD.
- Note that it is a very high-level storage brief design, for Neo is still developing. If more details are preferred, you need to dive into the source code.

## Terms

- Model: Casbin model.
- Schema: Derived from the model.
- Row Handle: A logical row handler based on several column files.
- Column: A attribute of schema, owns a column data file and a manifest file.
- Block: Basic disk RW UNIT, 4KB.
- Metadata: Schema metadata, owns the meta-information about each column’s manifest.
- Manifest: column meta-file, containing version information, etc.

## Overview

![overview.png](https://s2.loli.net/2022/06/19/9qnthUywiXe6lS3.png)

## Metadata File

### Metadata

Metadata is owned by a Schema, which represents the whole state of the schema at a certain moment.

Such as:

- Schema Information
- Schema Level WAL
- Manifests Information
- Row Handle Information

### Manifest

Manifest is owned by a Column, which represents the whole state of the column at a certain moment.

Such as:

- SST File Information
- Sequence Number

## Data File Layout

![block.png](https://s2.loli.net/2022/06/19/eD7U4m6HTaEJYtX.png)

SST file owns a set of blocks and a footer. The footer contains the metadata, like Index Block and Data Block area's position.

### Block

Block size is 4KB, and records are not allowed to skip blocks to store data, which also means the record has a **limitation** in its size.

- Data Block

  The data block consists of records and a header. 

  ```jsx
  |Header| Record(len + data) | Record(len + data) | ...|
  ```

- Index Block

  Considering reducing the duplicated data on disk, we are going to support two basic indexes.

  - bloom filter
  - min-max index

### Record

As Neo is a casbin compatible column storage, its record is actually a small variable-length binary stream.

## Data Path

CURD in LSM is actually RW operations with a background Compaction.

RW operation is based on RowSet. RowSet will dispatch to its columns. Compaction is column independent.

### Write

Before we write a policy, like `p = sub, obj, act`, we need to specificy a **primary key** in order to sort policy in the memory. And when the in-memory data dumps to disk, the row order remains.

> Write Path:  WAL → Memory → Disk,

### Read

In a schema like `p = sub, obj, act`, if we try to read it from disk in equality fields, we first use the primary key to get the row number of the record. And then use the row number to get other column value then dicide if the policy matches.

> Currently we are only consider to support primary. A seperate secondary index will be introduce later on.

### Compaction

Compaction is independent at the column level. So we can take a look at leveldb’s compaction design.

At common cases, compaction will cause LSM Version change, which may lead to manifest metadata change, so every time after column LSM compaction, we may need to change Metadata File of the Schema.
