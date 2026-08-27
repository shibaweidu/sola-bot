UPDATE point_center_configs
SET invite_success_text = E'🎉 入群成功！\n\n你已获得：+{invitee_points} 积分\n当前积分：{points}\n\n积分可兑换额度，免费领取更多权益。'
WHERE position(chr(92) || 'n' in invite_success_text) > 0;

UPDATE point_center_configs
SET points_text = '邀请好友获得积分，积分可兑换额度。这里查看当前积分、邀请奖励和免费额度兑换记录。'
WHERE points_text = ''
   OR points_text = '查看当前积分、今日邀请奖励和历史兑换额度。';
