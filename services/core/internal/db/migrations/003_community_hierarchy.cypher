-- Add parent_id to support hierarchical community structure.
-- Neo4j properties are schema-less, so existing Community nodes implicitly have
-- a missing parent_id. The repository treats a missing property as NULL.
-- This migration only adds the supporting index for parent lookups.

CREATE INDEX community_parent_id IF NOT EXISTS
FOR (c:Community) ON (c.parent_id);
