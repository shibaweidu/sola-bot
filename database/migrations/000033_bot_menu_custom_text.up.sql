ALTER TABLE bot_menu_items
    ADD COLUMN IF NOT EXISTS message_text TEXT NOT NULL DEFAULT '';

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conrelid = 'bot_menu_items'::regclass
          AND conname = 'bot_menu_items_action_type_check'
    ) THEN
        ALTER TABLE bot_menu_items DROP CONSTRAINT bot_menu_items_action_type_check;
    END IF;
    ALTER TABLE bot_menu_items
        ADD CONSTRAINT bot_menu_items_action_type_check CHECK (action_type IN ('builtin', 'link', 'custom'));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;
