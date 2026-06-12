-- Constraints
CREATE CONSTRAINT concept_id IF NOT EXISTS
FOR (c:Concept) REQUIRE c.id IS UNIQUE;

CREATE CONSTRAINT community_id IF NOT EXISTS
FOR (c:Community) REQUIRE c.id IS UNIQUE;

-- Indexes
CREATE INDEX concept_name IF NOT EXISTS FOR (c:Concept) ON (c.name);
CREATE INDEX concept_status IF NOT EXISTS FOR (c:Concept) ON (c.status);
CREATE INDEX concept_first_appeared IF NOT EXISTS FOR (c:Concept) ON (c.first_appeared);
CREATE INDEX community_level IF NOT EXISTS FOR (c:Community) ON (c.level);
CREATE INDEX community_domain IF NOT EXISTS FOR (c:Community) ON (c.domain);
