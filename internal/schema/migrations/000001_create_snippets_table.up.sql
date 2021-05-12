CREATE TABLE snippets
(
    snippet_id 	  UUID,
    title 		  TEXT,
    content		  TEXT,
    date_expires  TIMESTAMP WITH TIME ZONE,
    date_created  TIMESTAMP WITH TIME ZONE,
    date_updated  TIMESTAMP WITH TIME ZONE,
    PRIMARY KEY (snippet_id)
);
