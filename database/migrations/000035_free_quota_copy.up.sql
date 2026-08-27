UPDATE point_center_configs
SET invite_text = '🎁 免费领取额度！邀请好友加入群组，成功后你将获得积分；积分可以兑换额度，邀请越多，免费额度越多。'
WHERE invite_text IN (
    '邀请好友启动机器人并完成入群验证，邀请人和被邀请人都可获得积分。',
    '邀请好友启动机器人并完成入群验证，双方都可获得积分。'
);

UPDATE point_center_configs
SET invite_join_text = '请加入目标群组，完成频道订阅和入群验证。验证成功后邀请关系生效，积分可用于兑换免费额度。'
WHERE invite_join_text = '';

UPDATE point_center_configs
SET invite_success_text = E'🎉 入群成功！\n\n你已获得：+{invitee_points} 积分\n当前积分：{points}\n\n积分可兑换额度，免费领取更多权益。'
WHERE invite_success_text = ''
   OR invite_success_text = E'🎉 恭喜你已成功加入「{group}」！\n\n本次邀请奖励已到账：+{invitee_points} 积分\n当前积分：{points}';

UPDATE point_center_configs
SET points_text = '邀请好友获得积分，积分可兑换额度。这里查看当前积分、邀请奖励和免费额度兑换记录。'
WHERE points_text = ''
   OR points_text = '查看当前积分、今日邀请奖励和历史兑换额度。';

UPDATE point_center_configs
SET exchange_text = '🎁 用邀请获得的积分兑换免费额度。1 个积分兑换 1 个额度，最低兑换 10 个积分。'
WHERE exchange_text = ''
   OR exchange_text = '1 个积分兑换 1 个额度，最低兑换 10 个积分。';

UPDATE bot_menu_items
SET label = '🎁 邀请免费领额度', icon = '🎁', updated_at = NOW()
WHERE action_key = 'invite_rewards' AND label = '🎁 邀请获得积分';

UPDATE bot_menu_items
SET label = '🎫 积分兑换免费额度', icon = '🎫', updated_at = NOW()
WHERE action_key = 'exchange' AND label = '🎫 积分兑换额度';
