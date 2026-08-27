UPDATE point_center_configs
SET invite_join_text = '请加入目标群组，完成频道订阅和入群验证。验证成功后邀请关系生效，积分可用于兑换免费额度。'
WHERE invite_join_text LIKE '🚀 请先加入目标群组%';
