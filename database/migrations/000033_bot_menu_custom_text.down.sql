ALTER TABLE bot_menu_items DROP COLUMN IF EXISTS message_text;

-- Custom actions cannot be represented by the pre-000033 schema. Keep the
-- button rows while falling back to the harmless built-in help action.
UPDATE bot_menu_items
SET action_type = 'builtin', action_key = 'help', action_value = ''
WHERE action_type = 'custom';

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
        ADD CONSTRAINT bot_menu_items_action_type_check CHECK (action_type IN ('builtin', 'link'));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;
