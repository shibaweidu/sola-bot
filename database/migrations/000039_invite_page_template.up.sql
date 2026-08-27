ALTER TABLE point_center_configs
    ADD COLUMN IF NOT EXISTS invite_page_template TEXT NOT NULL DEFAULT '';

UPDATE point_center_configs
SET invite_page_template = E'🎁 邀请免费领额度\n━━━━━━━━━━\n\n邀请好友加入群组，完成入群验证后，你将获得积分。\n积分可以兑换额度，无需直接购买，邀请越多，免费额度越多。\n\n邀请人获得：{inviter_reward} 积分\n被邀请人获得：{invitee_reward} 积分\n\n第一步：把下面的机器人邀请链接分享给好友\n{invite_link}\n\n第二步：好友加入「{group}」，完成频道订阅和入群验证。\n\n验证成功后，双方均可获得积分，积分可兑换免费额度。\n兑换比例：{exchange_rate} 积分 = 1 个额度\n最低兑换：{exchange_minimum} 积分'
WHERE invite_page_template = '';
