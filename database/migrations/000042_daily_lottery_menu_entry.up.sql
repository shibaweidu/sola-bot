DO $$
DECLARE
    menu_row RECORD;
    next_row INTEGER;
    label_value TEXT;
    suffix INTEGER;
BEGIN
    FOR menu_row IN
        SELECT DISTINCT telegram_bot_id, role
        FROM bot_menu_items
    LOOP
        IF NOT EXISTS (
            SELECT 1
            FROM bot_menu_items
            WHERE telegram_bot_id = menu_row.telegram_bot_id
              AND role = menu_row.role
              AND action_type = 'builtin'
              AND action_key = 'daily_lottery'
        ) THEN
            label_value := '🎲 每日额度抽奖';
            suffix := 2;
            WHILE EXISTS (
                SELECT 1
                FROM bot_menu_items
                WHERE telegram_bot_id = menu_row.telegram_bot_id
                  AND role = menu_row.role
                  AND label = label_value
            ) LOOP
                label_value := format('🎲 每日额度抽奖 %s', suffix);
                suffix := suffix + 1;
            END LOOP;

            SELECT COALESCE(MAX(row_index), -1) + 1
            INTO next_row
            FROM bot_menu_items
            WHERE telegram_bot_id = menu_row.telegram_bot_id
              AND role = menu_row.role;

            INSERT INTO bot_menu_items (
                telegram_bot_id, role, label, icon, action_type, action_key,
                row_index, column_index, enabled
            ) VALUES (
                menu_row.telegram_bot_id, menu_row.role, label_value, '🎲', 'builtin', 'daily_lottery',
                next_row, 0, TRUE
            );
        END IF;
    END LOOP;
END $$;
