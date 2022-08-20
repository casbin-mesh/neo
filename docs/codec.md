# Codec

## Meta Codec
The Meta codec maps the name defined by the user(db name, table name, etc.)  to the globally unique id. 

For now, we use the following keys to get unique monotonically increasing global ids (id start from 1):

- `g_NextGlobalID_db`: Next db id 
- `g_NextGlobalID_table`: Next table id
- `g_NextGlobalID_index`: Next index id
- `g_NextGlobalID_matcher`: Next matcher id
- `g_NextGlobalID_column`: Next column id

And each id will store in the corresponding keys.

- `m_n{db_name}`: db_id
- `m_d{db_id}_t{table_name}`: table_id
- `m_d{db_id}_m{matcher_name}`: matcher_id
- `m_t{table_id}_i{index_name}`: index_id
- `m_t{table_id}_c{column_name}`: column_id


## Schema Model Codec

We store schema model information in the following keys:

- `s_d{db_id}` : db model
- `s_t{table_id}` : table model
- `s_c{column_id}` : column model
- `s_c{index_id}` : index model
- `s_m{index_id}` : matcher model

## Tuple Codec

Each row(tuple) in the table was stored in the following keys:

`t{table_id}_r{r_id}`

The `r_id` can identify a row in the table. In other words, it's unique in a table.

## Index Codec

We use the following format to store index info.

### Primary index Codec

```
Key: `i{index_id}_{index_column_value}`
Value: `r_id`
```

### Secondary(non-clustered) index Codec

```
Key: `i{index_id}_{leftmost_column_value}_{r_id}`
Value: `index_columns_value`
```

Due to each `index_id` being unique in global, we can store info in the above keys.