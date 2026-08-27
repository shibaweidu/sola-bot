UPDATE point_center_configs
SET exchange_instructions = '打开兑换网站，输入兑换码即可完成免费额度兑换。'
WHERE exchange_instructions = ''
   OR exchange_instructions = '打开兑换网站，输入兑换码即可完成兑换。';

UPDATE point_center_configs
SET invite_text = E'🎁 免费领取额度\n\n邀请好友加入群组，完成入群验证后，你将获得积分。\n积分可以兑换额度，无需直接购买，邀请越多，免费额度越多。'
WHERE invite_text IN (
    '🎁 免费领取额度！邀请好友加入群组，成功后你将获得积分；积分可以兑换额度，邀请越多，免费额度越多。',
    '邀请好友启动机器人并完成入群验证，邀请人和被邀请人都可获得积分。',
    '邀请好友启动机器人并完成入群验证，双方都可获得积分。'
);

UPDATE point_center_configs
SET invite_join_text = E'🚀 请先加入目标群组\n\n完成频道订阅和入群验证后，才能获得积分。\n\n验证成功后，您和邀请人都将获得积分，积分可兑换免费额度。'
WHERE invite_join_text IN (
    '',
    '请加入目标群组，完成频道订阅和入群验证。验证成功后邀请关系生效，积分可用于兑换免费额度。'
);

UPDATE point_center_configs
SET invite_success_text = E'🎉 入群成功！\n\n您和邀请人均可获得积分。\n本次获得：+{invitee_points} 积分\n当前积分：{points}\n\n积分可兑换免费额度。'
WHERE invite_success_text = ''
   OR invite_success_text = E'🎉 入群成功！\n\n你已获得：+{invitee_points} 积分\n当前积分：{points}\n\n积分可兑换额度，免费领取更多权益。';

UPDATE point_center_configs
SET points_text = '邀请好友获得积分，积分可兑换免费额度。这里查看当前积分、邀请奖励和兑换记录。'
WHERE points_text IN (
    '',
    '查看当前积分、今日邀请奖励和历史兑换额度。',
    '邀请好友获得积分，积分可兑换额度。这里查看当前积分、邀请奖励和免费额度兑换记录。'
);

UPDATE point_center_configs
SET sign_text = '每天签到可获得积分，积分可用于兑换免费额度。'
WHERE sign_text = '';

UPDATE point_center_configs
SET exchange_text = '🎁 使用邀请获得的积分兑换免费额度。1 个积分兑换 1 个额度，最低兑换 10 个积分。'
WHERE exchange_text IN (
    '',
    '1 个积分兑换 1 个额度，最低兑换 10 个积分。',
    '🎁 用邀请获得的积分兑换免费额度。1 个积分兑换 1 个额度，最低兑换 10 个积分。'
);

UPDATE point_center_configs
SET purchase_text = '需要更多额度？可以直接购买额度。'
WHERE purchase_text = '';

UPDATE point_center_configs
SET shop_text = '浏览小铺商品，购买更多额度和相关服务。'
WHERE shop_text = '';

UPDATE point_center_configs
SET rank_text = '邀请人数最多的前 5 名用户，邀请越多，获得的免费额度越多。'
WHERE rank_text IN ('', '邀请人数最多的前 5 名用户。');
