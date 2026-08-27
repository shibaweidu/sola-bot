UPDATE point_center_configs
SET invite_text = '邀请好友启动机器人并完成入群验证，双方都可获得积分。'
WHERE invite_text = '🎁 免费领取额度！邀请好友加入群组，成功后你将获得积分；积分可以兑换额度，邀请越多，免费额度越多。';

UPDATE bot_menu_items
SET label = '🎁 邀请获得积分', icon = '🎁', updated_at = NOW()
WHERE action_key = 'invite_rewards' AND label = '🎁 邀请免费领额度';

UPDATE bot_menu_items
SET label = '🎫 积分兑换额度', icon = '🎫', updated_at = NOW()
WHERE action_key = 'exchange' AND label = '🎫 积分兑换免费额度';
