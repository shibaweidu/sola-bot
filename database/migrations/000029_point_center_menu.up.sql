CREATE TEMP TABLE legacy_member_menu_bots AS
SELECT telegram_bot_id
FROM bot_menu_items
WHERE role = 'member'
GROUP BY telegram_bot_id
HAVING COUNT(*) = 6
   AND COUNT(*) FILTER (WHERE (action_key, label) IN (('points', '💎 我的积分'), ('sign', '📅 每日签到'), ('rank', '🏆 排行榜'), ('lottery', '🎁 抽奖大厅'), ('help', '❓ 使用帮助'), ('info', 'ℹ️ 会话信息'))) = 6;

UPDATE bot_menu_items item
SET label = CASE item.action_key WHEN 'points' THEN '💎 我的积分' WHEN 'sign' THEN '📅 每日签到' WHEN 'rank' THEN '🏆 积分榜' END,
    icon = CASE item.action_key WHEN 'points' THEN '💎' WHEN 'sign' THEN '📅' WHEN 'rank' THEN '🏆' END,
    row_index = CASE item.action_key WHEN 'points' THEN 0 WHEN 'sign' THEN 1 WHEN 'rank' THEN 3 END,
    column_index = CASE item.action_key WHEN 'points' THEN 1 WHEN 'sign' THEN 0 WHEN 'rank' THEN 0 END,
    updated_at = NOW()
FROM legacy_member_menu_bots legacy
WHERE item.telegram_bot_id = legacy.telegram_bot_id AND item.role = 'member' AND item.action_key IN ('points', 'sign', 'rank');

DELETE FROM bot_menu_items item USING legacy_member_menu_bots legacy
WHERE item.telegram_bot_id = legacy.telegram_bot_id AND item.role = 'member' AND item.action_key IN ('lottery', 'help', 'info');

INSERT INTO bot_menu_items (telegram_bot_id, role, label, icon, action_type, action_key, action_value, row_index, column_index, enabled)
SELECT legacy.telegram_bot_id, item.role, item.label, item.icon, 'builtin', item.action_key, '', item.row_index, item.column_index, TRUE
FROM legacy_member_menu_bots legacy
CROSS JOIN (VALUES
    ('member', '🎁 邀请获得积分', '🎁', 'invite_rewards', 0, 0),
    ('member', '🎫 积分兑换额度', '🎫', 'exchange', 1, 1),
    ('member', '🛒 直接购买额度', '🛒', 'purchase', 2, 0),
    ('member', '🏪 小铺地址', '🏪', 'shop', 2, 1)
) AS item(role, label, icon, action_key, row_index, column_index);

CREATE TEMP TABLE legacy_admin_menu_bots AS
SELECT telegram_bot_id
FROM bot_menu_items
WHERE role = 'admin'
GROUP BY telegram_bot_id
HAVING COUNT(*) = 10
   AND COUNT(*) FILTER (WHERE (action_key, label) IN (('points', '💎 我的积分'), ('sign', '📅 每日签到'), ('rank', '🏆 排行榜'), ('lottery', '🎁 抽奖大厅'), ('help', '❓ 使用帮助'), ('info', 'ℹ️ 会话信息'), ('private_console', '📋 运营工作台'), ('admin_center', '🛡 群管中心'), ('admin_config', '⚙️ 群组配置'), ('scheduled_posts', '📣 定时发帖'))) = 10;

UPDATE bot_menu_items item
SET label = CASE item.action_key WHEN 'points' THEN '💎 我的积分' WHEN 'sign' THEN '📅 每日签到' WHEN 'rank' THEN '🏆 积分榜' WHEN 'private_console' THEN '📋 运营工作台' WHEN 'admin_center' THEN '🛡 群管中心' WHEN 'admin_config' THEN '⚙️ 群组配置' WHEN 'scheduled_posts' THEN '📣 定时发帖' END,
    icon = CASE item.action_key WHEN 'points' THEN '💎' WHEN 'sign' THEN '📅' WHEN 'rank' THEN '🏆' WHEN 'private_console' THEN '📋' WHEN 'admin_center' THEN '🛡' WHEN 'admin_config' THEN '⚙️' WHEN 'scheduled_posts' THEN '📣' END,
    row_index = CASE item.action_key WHEN 'points' THEN 0 WHEN 'sign' THEN 1 WHEN 'rank' THEN 3 WHEN 'private_console' THEN 4 WHEN 'admin_center' THEN 4 WHEN 'admin_config' THEN 5 WHEN 'scheduled_posts' THEN 5 END,
    column_index = CASE item.action_key WHEN 'points' THEN 1 WHEN 'sign' THEN 0 WHEN 'rank' THEN 0 WHEN 'private_console' THEN 0 WHEN 'admin_center' THEN 1 WHEN 'admin_config' THEN 0 WHEN 'scheduled_posts' THEN 1 END,
    updated_at = NOW()
FROM legacy_admin_menu_bots legacy
WHERE item.telegram_bot_id = legacy.telegram_bot_id AND item.role = 'admin' AND item.action_key IN ('points', 'sign', 'rank', 'private_console', 'admin_center', 'admin_config', 'scheduled_posts');

DELETE FROM bot_menu_items item USING legacy_admin_menu_bots legacy
WHERE item.telegram_bot_id = legacy.telegram_bot_id AND item.role = 'admin' AND item.action_key IN ('lottery', 'help', 'info');

INSERT INTO bot_menu_items (telegram_bot_id, role, label, icon, action_type, action_key, action_value, row_index, column_index, enabled)
SELECT legacy.telegram_bot_id, item.role, item.label, item.icon, 'builtin', item.action_key, '', item.row_index, item.column_index, TRUE
FROM legacy_admin_menu_bots legacy
CROSS JOIN (VALUES
    ('admin', '🎁 邀请获得积分', '🎁', 'invite_rewards', 0, 0),
    ('admin', '🎫 积分兑换额度', '🎫', 'exchange', 1, 1),
    ('admin', '🛒 直接购买额度', '🛒', 'purchase', 2, 0),
    ('admin', '🏪 小铺地址', '🏪', 'shop', 2, 1)
) AS item(role, label, icon, action_key, row_index, column_index);

DROP TABLE legacy_member_menu_bots;
DROP TABLE legacy_admin_menu_bots;
