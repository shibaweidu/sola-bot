<template>
  <div class="page">
    <PageHeader
      eyebrow="积分运营"
      title="积分中心"
      description="配置邀请得积分、积分兑换免费额度、兑换地址和兑换码库存。"
    >
      <template #actions>
        <ChatSelect v-model="selectedChatId" @update:model-value="loadConfig" />
        <el-button :icon="Refresh" :loading="loading" @click="loadAll">刷新</el-button>
        <el-button type="primary" :icon="Check" :loading="saving" @click="saveConfig">保存配置</el-button>
      </template>
    </PageHeader>

    <el-tabs v-model="activeTab" type="border-card">
      <el-tab-pane label="规则与文案" name="config">
        <el-form :model="form" label-position="top" class="config-form">
          <el-row :gutter="16">
            <el-col :xs="24" :md="12">
              <PanelSection title="奖励规则" description="规则按当前选中的群组生效。">
                <div class="switch-grid">
                  <div class="switch-row"><span>邀请奖励</span><el-switch v-model="form.invite_enabled" /></div>
                  <div class="switch-row"><span>每日签到</span><el-switch v-model="form.sign_enabled" /></div>
                  <div class="switch-row"><span>积分兑换</span><el-switch v-model="form.exchange_enabled" /></div>
                </div>
                <div class="number-grid">
                  <el-form-item label="邀请人奖励"><el-input-number v-model="form.inviter_reward" :min="0" /></el-form-item>
                  <el-form-item label="被邀请人奖励"><el-input-number v-model="form.invitee_reward" :min="0" /></el-form-item>
                  <el-form-item label="签到奖励"><el-input-number v-model="form.sign_reward" :min="0" /></el-form-item>
                  <el-form-item label="最低兑换积分"><el-input-number v-model="form.exchange_minimum" :min="1" /></el-form-item>
                  <el-form-item label="兑换比例（额度/积分）"><el-input-number v-model="form.exchange_rate" :min="1" /></el-form-item>
                </div>
              </PanelSection>
            </el-col>
            <el-col :xs="24" :md="12">
              <PanelSection title="地址设置" description="仅允许 HTTPS 地址。">
                <el-form-item label="兑换网站地址（可选）"><el-input v-model="form.exchange_url" placeholder="留空则只显示兑换说明" /></el-form-item>
                <el-form-item label="直接购买地址（可选）"><el-input v-model="form.purchase_url" placeholder="留空则只显示购买文案" /></el-form-item>
                <el-form-item label="小铺地址（可选）"><el-input v-model="form.shop_url" placeholder="留空则只显示小铺文案" /></el-form-item>
                <el-form-item label="兑换方式说明"><el-input v-model="form.exchange_instructions" type="textarea" :rows="3" /></el-form-item>
              </PanelSection>
            </el-col>
          </el-row>
              <PanelSection title="用户看到的文案" description="留空时使用系统默认文案。">
            <div class="copy-grid">
              <el-form-item label="分享邀请文案">
                <el-input v-model="form.invite_text" type="textarea" :rows="3" />
                <div class="field-hint">用于 Telegram 分享窗口的短文案，不影响下方邀请页面正文。</div>
              </el-form-item>
              <el-form-item label="邀请页面完整模板" class="invite-template-item">
                <el-input v-model="form.invite_page_template" type="textarea" :rows="10" maxlength="4000" show-word-limit placeholder="例如：邀请链接：{invite_link}\n目标群组：{group}" />
                <div class="field-hint">这里的内容会完整替换机器人“邀请获得积分”页面中的正文，不会再额外拼接第一步、第二步等固定文字。</div>
                <div class="template-vars">
                  <div>可用变量：</div>
                  <span><code>{invite_link}</code> 邀请链接</span>
                  <span><code>{group}</code> 目标群组名称</span>
                  <span><code>{inviter_reward}</code> 邀请人奖励积分</span>
                  <span><code>{invitee_reward}</code> 被邀请人奖励积分</span>
                  <span><code>{exchange_minimum}</code> 最低兑换积分</span>
                  <span><code>{exchange_rate}</code> 兑换比例（额度/积分）</span>
                  <span><code>{invite_join_url}</code> 目标群组入群链接</span>
                </div>
                <div class="template-preview">
                  <div class="template-preview-title">消息预览</div>
                  <pre>{{ inviteTemplatePreview }}</pre>
                </div>
              </el-form-item>
              <el-form-item label="目标群组入群链接"><el-input v-model="form.invite_join_url" placeholder="https://t.me/your_group" /></el-form-item>
              <el-form-item label="入群提示文案"><el-input v-model="form.invite_join_text" type="textarea" :rows="3" /></el-form-item>
              <el-form-item label="入群成功文案">
                <el-input v-model="form.invite_success_text" type="textarea" :rows="4" placeholder="支持 {name}、{group}、{invitee_points}、{points}" />
                <div class="field-hint">变量：<code>{name}</code> 用户名；<code>{group}</code> 群组名称；<code>{invitee_points}</code> 本次获得积分；<code>{points}</code> 当前积分。</div>
              </el-form-item>
              <el-form-item label="我的积分"><el-input v-model="form.points_text" type="textarea" :rows="3" /></el-form-item>
              <el-form-item label="签到页面"><el-input v-model="form.sign_text" type="textarea" :rows="3" /></el-form-item>
              <el-form-item label="积分兑换免费额度文案"><el-input v-model="form.exchange_text" type="textarea" :rows="3" /></el-form-item>
              <el-form-item label="购买页面文案"><el-input v-model="form.purchase_text" type="textarea" :rows="3" /></el-form-item>
              <el-form-item label="小铺页面文案"><el-input v-model="form.shop_text" type="textarea" :rows="3" /></el-form-item>
              <el-form-item label="积分榜页面"><el-input v-model="form.rank_text" type="textarea" :rows="3" /></el-form-item>
            </div>
          </PanelSection>
          <PanelSection title="测试工具" description="仅用于测试环境，生产环境请谨慎使用。">
            <div class="adjust-tool">
              <div class="adjust-fields">
                <el-form-item label="Telegram 用户 ID">
                  <el-input v-model="adjustUserId" placeholder="例如：8720946378" />
                </el-form-item>
                <el-form-item label="积分变化">
                  <el-input-number v-model="adjustDelta" :min="-999999" :max="999999" />
                </el-form-item>
                <el-form-item label="调整原因">
                  <el-input v-model="adjustReason" maxlength="64" placeholder="例如：测试邀请奖励" />
                </el-form-item>
              </div>
              <div class="adjust-actions">
                <el-button type="primary" :loading="adjusting" @click="adjustPoints">确认调整积分</el-button>
                <span class="adjust-hint">支持正数加分和负数扣分，操作会写入积分日志。</span>
              </div>
              <div v-if="adjustResult" class="adjust-result">
                调整前 {{ adjustResult.before }} 分，{{ adjustResult.delta >= 0 ? '+' : '' }}{{ adjustResult.delta }} 分，调整后 {{ adjustResult.after }} 分
              </div>
            </div>
            <div class="reset-tool">
              <el-input v-model="resetUserId" placeholder="输入 Telegram 用户 ID" />
              <el-button type="warning" :loading="resetting" @click="resetReferral">完全重置测试用户</el-button>
            </div>
            <div class="reset-hint">重置后，该用户下一次完成入群验证会再次按当前规则发放奖励，并重新发送成功私聊。生产环境请勿使用。</div>
          </PanelSection>
        </el-form>
      </el-tab-pane>

      <el-tab-pane label="兑换码库存" name="inventory">
        <el-row :gutter="16">
          <el-col :xs="24" :md="10">
            <PanelSection title="批量导入兑换码" description="每行一个兑换码，重复码会自动跳过。">
              <el-form label-position="top">
                <el-form-item label="批次名称"><el-input v-model="inventory.batch_name" placeholder="例如：2026-08-首批" /></el-form-item>
                <el-form-item label="每个兑换码对应额度"><el-input-number v-model="inventory.amount" :min="1" /><div class="field-hint">不同额度请分别导入，例如 10、50、100 额度分别建立库存。</div></el-form-item>
                <el-form-item label="兑换网站地址"><el-input v-model="inventory.redeem_url" placeholder="留空使用群组兑换地址" /></el-form-item>
                <el-form-item label="兑换码"><el-input v-model="inventory.codes" type="textarea" :rows="12" placeholder="CODE-001\nCODE-002\nCODE-003" /></el-form-item>
                <el-button type="primary" :loading="importing" @click="submitImport">导入库存</el-button>
              </el-form>
            </PanelSection>
          </el-col>
          <el-col :xs="24" :md="14">
            <PanelSection title="库存列表" description="兑换成功后兑换码会自动变为已分配。">
              <div v-if="summaries.length" class="inventory-summary-grid">
                <button v-for="item in summaries" :key="item.amount" type="button" class="inventory-summary" :class="{ active: inventory.amountFilter === item.amount }" @click="selectAmount(item.amount)">
                  <strong>{{ item.amount }} 额度</strong>
                  <span>可用 {{ item.available }} / 总量 {{ item.total }}</span>
                  <small>已分配 {{ item.assigned }} · 已使用 {{ item.used }}</small>
                </button>
                <button v-if="inventory.amountFilter !== undefined" type="button" class="clear-amount" @click="selectAmount(undefined)">查看全部额度</button>
              </div>
              <div class="inventory-toolbar">
                <el-select v-model="inventory.status" clearable placeholder="全部状态" @change="loadCodes">
                  <el-option label="可用" value="available" /><el-option label="已分配" value="assigned" /><el-option label="已使用" value="used" />
                </el-select>
                <el-button :icon="Refresh" @click="loadInventory">刷新库存</el-button>
              </div>
              <el-table :data="codes" stripe size="small" max-height="520">
                <el-table-column prop="code" label="兑换码" min-width="180" show-overflow-tooltip />
                <el-table-column prop="batch_name" label="批次" min-width="120" />
                <el-table-column prop="amount" label="额度" width="70" />
                <el-table-column prop="status" label="状态" width="90" />
                <el-table-column prop="redeem_url" label="兑换地址" min-width="150" show-overflow-tooltip />
              </el-table>
            </PanelSection>
          </el-col>
        </el-row>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { Check, Refresh } from "@element-plus/icons-vue";
import ChatSelect from "@/components/ChatSelect.vue";
import PageHeader from "@/components/PageHeader.vue";
import PanelSection from "@/components/PanelSection.vue";
import { fetchExchangeCodeSummary, fetchExchangeCodes, fetchPointCenterConfig, importExchangeCodes, resetReferralForTesting, updatePointCenterConfig, type ExchangeCodeRecord, type ExchangeCodeSummary, type PointCenterConfig } from "@/api/pointCenter";
import { fetchUserPointDetail, updateUserPoints } from "@/api/points";

const selectedChatId = ref("");
const activeTab = ref("config");
const loading = ref(false);
const saving = ref(false);
const importing = ref(false);
const resetting = ref(false);
const resetUserId = ref("");
const adjusting = ref(false);
const adjustUserId = ref("");
const adjustDelta = ref(10);
const adjustReason = ref("");
const adjustResult = ref<{ before: number; delta: number; after: number }>();
const codes = ref<ExchangeCodeRecord[]>([]);
const summaries = ref<ExchangeCodeSummary[]>([]);
const form = reactive<PointCenterConfig>({ chat_id: 0, invite_enabled: true, inviter_reward: 100, invitee_reward: 20, sign_enabled: true, sign_reward: 1, exchange_enabled: true, exchange_minimum: 10, exchange_rate: 1, exchange_url: "", exchange_instructions: "", purchase_url: "", purchase_text: "", shop_url: "", shop_text: "", invite_text: "", invite_page_template: "", invite_join_url: "", invite_join_text: "", invite_success_text: "", points_text: "", sign_text: "", exchange_text: "", rank_text: "" });
const inventory = reactive({ batch_name: "", amount: 10, redeem_url: "", codes: "", status: "", amountFilter: undefined as number | undefined });
const inviteTemplatePreview = computed(() => {
  const source = form.invite_page_template.trim() || "🎁 邀请免费领额度\n━━━━━━━━━━\n\n邀请好友加入群组，完成入群验证后，你将获得积分。\n积分可以兑换额度，无需直接购买，邀请越多，免费额度越多。\n\n邀请人获得：{inviter_reward} 积分\n被邀请人获得：{invitee_reward} 积分\n\n第一步：把下面的机器人邀请链接分享给好友\n{invite_link}\n\n第二步：好友加入「{group}」，完成频道订阅和入群验证。\n\n验证成功后，双方均可获得积分，积分可兑换免费额度。\n兑换比例：{exchange_rate} 积分 = 1 个额度\n最低兑换：{exchange_minimum} 积分";
  return source
    .replaceAll("{invite_link}", "https://t.me/你的机器人?start=ref_demo")
    .replaceAll("{group}", "示例群组")
    .replaceAll("{inviter_reward}", String(form.inviter_reward))
    .replaceAll("{invitee_reward}", String(form.invitee_reward))
    .replaceAll("{exchange_minimum}", String(form.exchange_minimum))
    .replaceAll("{exchange_rate}", String(form.exchange_rate))
    .replaceAll("{invite_join_url}", form.invite_join_url || "（未配置）");
});

async function loadConfig(): Promise<void> {
  if (!selectedChatId.value) return;
  loading.value = true;
  try { Object.assign(form, await fetchPointCenterConfig(selectedChatId.value)); } catch { ElMessage.error("积分中心配置加载失败"); } finally { loading.value = false; }
}
async function loadCodes(): Promise<void> {
  try { codes.value = (await fetchExchangeCodes(inventory.status, inventory.amountFilter)).items; } catch { ElMessage.error("兑换码库存加载失败"); }
}
async function loadInventory(): Promise<void> {
  try {
    const result = await Promise.all([fetchExchangeCodes(inventory.status, inventory.amountFilter), fetchExchangeCodeSummary()]);
    codes.value = result[0].items;
    summaries.value = result[1].items;
  } catch { ElMessage.error("兑换码库存加载失败"); }
}
function selectAmount(amount: number | undefined): void {
  inventory.amountFilter = amount;
  void loadCodes();
}
async function loadAll(): Promise<void> { await Promise.all([loadConfig(), loadInventory()]); }
async function saveConfig(): Promise<void> {
  if (!selectedChatId.value) { ElMessage.warning("请先选择群组"); return; }
  saving.value = true;
  try { Object.assign(form, await updatePointCenterConfig({ ...form, chat_id: Number(selectedChatId.value) })); ElMessage.success("积分中心配置已保存"); } catch { ElMessage.error("积分中心配置保存失败"); } finally { saving.value = false; }
}
async function resetReferral(): Promise<void> {
  const userId = Number(resetUserId.value.trim());
  if (!selectedChatId.value) { ElMessage.warning("请先选择群组"); return; }
  if (!Number.isSafeInteger(userId) || userId <= 0) { ElMessage.warning("请输入有效的 Telegram 用户 ID"); return; }
  try {
    await ElMessageBox.confirm("该操作会清空该用户在当前群组的积分、积分日志、签到和邀请测试状态，是否继续？", "确认完全重置", { type: "warning" });
  } catch { return; }
  resetting.value = true;
  try {
    await resetReferralForTesting(selectedChatId.value, userId);
    ElMessage.success("邀请测试状态已重置");
  } catch { ElMessage.error("测试用户重置失败，请检查用户 ID 和当前群组"); } finally { resetting.value = false; }
}
async function adjustPoints(): Promise<void> {
  if (!selectedChatId.value) { ElMessage.warning("请先选择群组"); return; }
  const userId = Number(adjustUserId.value.trim());
  if (!Number.isSafeInteger(userId) || userId <= 0) { ElMessage.warning("请输入有效的 Telegram 用户 ID"); return; }
  if (!Number.isSafeInteger(adjustDelta.value) || adjustDelta.value === 0) { ElMessage.warning("积分变化不能为 0"); return; }
  const reason = adjustReason.value.trim();
  if (!reason) { ElMessage.warning("请填写调整原因"); return; }
  try {
    await ElMessageBox.confirm(`将为用户 ${userId} 调整 ${adjustDelta.value > 0 ? "+" : ""}${adjustDelta.value} 积分，是否继续？`, "确认测试调分", { type: "warning" });
  } catch { return; }
  adjusting.value = true;
  try {
    const before = await fetchUserPointDetail(selectedChatId.value, userId);
    const after = await updateUserPoints(selectedChatId.value, userId, { delta: adjustDelta.value, reason });
    adjustResult.value = { before: before.total_points, delta: adjustDelta.value, after: after.total_points };
    ElMessage.success(`积分已调整，当前余额 ${after.total_points}`);
  } catch { ElMessage.error("积分调整失败，请检查用户 ID 和当前群组"); } finally { adjusting.value = false; }
}
async function submitImport(): Promise<void> {
  const values = inventory.codes.split(/\r?\n/).map((item) => item.trim()).filter(Boolean);
  if (!values.length) { ElMessage.warning("请填写兑换码"); return; }
  importing.value = true;
  try { const result = await importExchangeCodes({ batch_name: inventory.batch_name, amount: inventory.amount, redeem_url: inventory.redeem_url || form.exchange_url, codes: values }); ElMessage.success(`导入 ${result.imported} 个，跳过 ${result.skipped} 个`); inventory.codes = ""; await loadInventory(); } catch { ElMessage.error("兑换码导入失败"); } finally { importing.value = false; }
}
onMounted(loadAll);
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 16px; }
.config-form { display: flex; flex-direction: column; gap: 16px; }
.switch-grid, .number-grid, .copy-grid { display: grid; gap: 12px; }
.switch-grid { grid-template-columns: repeat(3, minmax(0, 1fr)); margin-bottom: 16px; }
.number-grid, .copy-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
.switch-row { display: flex; align-items: center; justify-content: space-between; border: 1px solid var(--app-border); padding: 10px 12px; border-radius: 8px; }
.inventory-toolbar { display: flex; gap: 10px; margin-bottom: 12px; }
.inventory-summary-grid { display: flex; flex-wrap: wrap; gap: 8px; margin-bottom: 14px; }
.inventory-summary { display: flex; min-width: 138px; flex-direction: column; align-items: flex-start; gap: 3px; border: 1px solid var(--app-border); border-radius: 8px; background: var(--app-surface-2); padding: 9px 11px; color: var(--app-text); cursor: pointer; text-align: left; }
.inventory-summary:hover, .inventory-summary.active { border-color: var(--app-primary); }
.inventory-summary span, .inventory-summary small { color: var(--app-muted); font-size: 12px; }
.clear-amount { border: 0; background: transparent; color: var(--app-primary); cursor: pointer; font-size: 12px; }
.adjust-tool { border: 1px solid var(--app-border); border-radius: 8px; padding: 12px; }
.adjust-fields { display: grid; grid-template-columns: 1.2fr 0.8fr 1.2fr; gap: 12px; }
.adjust-actions { display: flex; align-items: center; gap: 12px; }
.adjust-hint, .adjust-result { color: var(--app-muted); font-size: 12px; }
.adjust-result { margin-top: 10px; color: var(--app-text); }
.reset-tool { display: flex; gap: 10px; align-items: center; max-width: 560px; }
.reset-tool .el-input { max-width: 280px; }
.reset-hint { margin-top: 8px; color: var(--app-muted); font-size: 12px; }
.field-hint { margin-top: 6px; color: var(--app-muted); font-size: 12px; line-height: 1.5; }
.template-vars { display: flex; flex-wrap: wrap; gap: 4px 12px; margin-top: 6px; color: var(--app-muted); font-size: 12px; line-height: 1.7; }
.template-vars > div { flex-basis: 100%; }
.template-vars code, .field-hint code { margin-right: 3px; color: var(--app-primary); }
.template-preview { margin-top: 10px; border: 1px solid var(--app-border); border-radius: 8px; background: var(--app-surface-2); padding: 10px 12px; }
.template-preview-title { margin-bottom: 6px; color: var(--app-muted); font-size: 12px; }
.template-preview pre { margin: 0; white-space: pre-wrap; overflow-wrap: anywhere; font: inherit; line-height: 1.55; }
@media (max-width: 800px) { .switch-grid, .number-grid, .copy-grid { grid-template-columns: 1fr; } }
@media (max-width: 800px) { .adjust-fields { grid-template-columns: 1fr; } }
@media (max-width: 560px) { .reset-tool { align-items: stretch; flex-direction: column; } .reset-tool .el-input { max-width: none; } }
</style>
