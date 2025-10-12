package database

// SELECT
//     -- Table/Schema details
//     c.table_schema,
//     c.table_name,
//     c.column_name,
//     c.ordinal_position,

//     -- Table Comment
//     obj_description( (c.table_schema || '.' || c.table_name)::regclass, 'pg_class' ) AS table_comment,

//     -- Column Comment
//     col_description( (c.table_schema || '.' || c.table_name)::regclass, c.ordinal_position ) AS column_comment

// FROM
//     information_schema.columns c
// WHERE
//     c.table_schema NOT IN ('pg_catalog', 'information_schema')
//     -- Optional: filter to a specific schema
//     -- AND c.table_schema = 'public'
// ORDER BY
//     c.table_schema,
//     c.table_name,
//     c.ordinal_position;