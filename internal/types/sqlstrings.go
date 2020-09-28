package types

// CreateTables - sql создания новой таблицы и индексов
const CreateTables string = `
	CREATE TABLE IF NOT EXISTS "subscribers" (
		"id" BIGSERIAL NOT NULL PRIMARY KEY,
		"link" TEXT NOT NULL,
		"user_email" TEXT NOT NULL,
		"price" BIGINT NOT NULL
	);

	CREATE INDEX IF NOT EXISTS subs_email_idx ON subscribers (user_email);
	CREATE INDEX IF NOT EXISTS subs_link_idx ON subscribers (link);
	CREATE INDEX IF NOT EXISTS subs_price_idx ON subscribers (price);
`
