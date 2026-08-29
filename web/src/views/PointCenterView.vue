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
            <PanelSection title="库存摘要" description="明细已移至分页管理抽屉，避免兑换码过多撑长页面。">
              <div v-if="summaries.length" class="inventory-summary-grid">
                <button v-for="item in summaries" :key="item.amount" type="button" class="inventory-summary" :class="{ active: inventory.amountFilter === item.amount }" @click="selectAmount(item.amount)">
                  <strong>{{ item.amount }} 额度</strong>
                  <span>可用 {{ item.available }} / 总量 {{ item.total }}</span>
                  <small>已分配 {{ item.assigned }} · 已使用 {{ item.used }}</small>
                </button>
                <button v-if="inventory.amountFilter !== undefined" type="button" class="clear-amount" @click="selectAmount(undefined)">查看全部额度</button>
              </div>
              <div class="inventory-summary-actions">
                <span class="field-hint">点击额度卡片可直接筛选明细。</span>
                <el-button type="primary" :icon="Edit" @click="openInventoryDrawer">管理库存</el-button>
              </div>
            </PanelSection>
          </el-col>
        </el-row>
      </el-tab-pane>
    </el-tabs>

    <el-drawer v-model="inventoryDrawer" title="兑换码库存" size="min(920px, 95vw)" destroy-on-close @closed="selectedCodeIds = []">
      <div class="drawer-toolbar">
        <el-select v-model="inventory.status" clearable placeholder="全部状态" @change="reloadInventoryDrawer">
          <el-option label="可用" value="available" /><el-option label="已分配" value="assigned" /><el-option label="已使用" value="used" />
        </el-select>
        <el-button :icon="Refresh" :loading="inventoryLoading" @click="reloadInventoryDrawer">刷新</el-button>
        <span class="drawer-total">共 {{ inventoryTotal }} 条</span>
      </div>
      <div class="selection-toolbar" v-if="selectedCodeIds.length">
        <strong>已选 {{ selectedCodeIds.length }} 条</strong>
        <el-button type="primary" :icon="Edit" @click="openBatchEdit">批量编辑</el-button>
        <el-button type="danger" :icon="Delete" @click="deleteSelectedCodes">批量删除</el-button>
        <span class="field-hint">仅可用兑换码可编辑或删除</span>
      </div>
      <el-table ref="codeTable" :data="codes" row-key="id" stripe size="small" class="inventory-table" @selection-change="onSelectionChange">
        <el-table-column type="selection" width="48" :selectable="isCodeSelectable" reserve-selection />
        <el-table-column prop="code" label="兑换码" min-width="190" show-overflow-tooltip>
          <template #default="{ row }"><span class="code-cell">{{ row.code }}</span><el-button text :icon="CopyDocument" aria-label="复制兑换码" @click="copyCode(row.code)" /></template>
        </el-table-column>
        <el-table-column prop="batch_name" label="批次" min-width="130" show-overflow-tooltip />
        <el-table-column prop="amount" label="额度" width="78" />
        <el-table-column prop="status" label="状态" width="90" />
        <el-table-column prop="redeem_url" label="兑换地址" min-width="180" show-overflow-tooltip />
      </el-table>
      <div class="drawer-pagination"><el-pagination v-model:current-page="inventoryPage" v-model:page-size="inventoryPageSize" :page-sizes="[20, 50, 100]" layout="total, sizes, prev, pager, next" :total="inventoryTotal" @current-change="loadCodes" @size-change="handleInventorySizeChange" /></div>
    </el-drawer>

    <el-dialog v-model="batchEditVisible" title="批量编辑库存" width="500px">
      <p class="field-hint">已选择 {{ selectedCodeIds.length }} 条，仅修改勾选的字段；空值会清空对应字段。</p>
      <el-form label-position="top">
        <el-checkbox v-model="batchForm.edit_redeem_url">修改兑换地址</el-checkbox>
        <el-input v-model="batchForm.redeem_url" :disabled="!batchForm.edit_redeem_url" placeholder="留空可清除，或填写 https://..." />
        <el-checkbox v-model="batchForm.edit_batch_name">修改批次名称</el-checkbox>
        <el-input v-model="batchForm.batch_name" :disabled="!batchForm.edit_batch_name" maxlength="128" placeholder="留空可清除" />
        <el-checkbox v-model="batchForm.edit_expires_at">修改有效期</el-checkbox>
        <el-input v-model="batchForm.expires_at" :disabled="!batchForm.edit_expires_at" placeholder="RFC3339，例如 2026-12-31T23:59:59Z；留空清除" />
      </el-form>
      <template #footer><el-button @click="batchEditVisible = false">取消</el-button><el-button type="primary" :loading="batchEditing" @click="submitBatchEdit">保存修改</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { Check, CopyDocument, Delete, Edit, Refresh } from "@element-plus/icons-vue";
import ChatSelect from "@/components/ChatSelect.vue";
import PageHeader from "@/components/PageHeader.vue";
import PanelSection from "@/components/PanelSection.vue";
import { batchDeleteExchangeCodes, batchUpdateExchangeCodes, fetchExchangeCodeSummary, fetchExchangeCodes, fetchPointCenterConfig, importExchangeCodes, resetReferralForTesting, updatePointCenterConfig, type ExchangeCodeRecord, type ExchangeCodeSummary, type PointCenterConfig } from "@/api/pointCenter";
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
const inventoryDrawer = ref(false);
const inventoryLoading = ref(false);
const inventoryPage = ref(1);
const inventoryPageSize = ref(20);
const inventoryTotal = ref(0);
const selectedCodeIds = ref<string[]>([]);
const batchEditVisible = ref(false);
const batchEditing = ref(false);
const batchForm = reactive({ edit_redeem_url: false, edit_batch_name: false, edit_expires_at: false, redeem_url: "", batch_name: "", expires_at: "" });
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
  if (!inventoryDrawer.value) return;
  inventoryLoading.value = true;
  try {
    const result = await fetchExchangeCodes(inventory.status, inventory.amountFilter, inventoryPage.value, inventoryPageSize.value);
    codes.value = result.items;
    inventoryTotal.value = result.total;
  } catch { ElMessage.error("兑换码库存加载失败"); }
  finally { inventoryLoading.value = false; }
}
async function loadInventory(): Promise<void> {
  try {
    summaries.value = (await fetchExchangeCodeSummary()).items;
  } catch { ElMessage.error("兑换码库存加载失败"); }
}
function selectAmount(amount: number | undefined): void {
  inventory.amountFilter = amount;
  inventoryPage.value = 1;
  if (inventoryDrawer.value) void loadCodes();
}
function openInventoryDrawer(): void { inventoryDrawer.value = true; inventoryPage.value = 1; void loadCodes(); }
function reloadInventoryDrawer(): void { inventoryPage.value = 1; selectedCodeIds.value = []; void loadCodes(); }
function handleInventorySizeChange(size: number): void { inventoryPageSize.value = size; inventoryPage.value = 1; void loadCodes(); }
function isCodeSelectable(row: ExchangeCodeRecord): boolean { return row.status === "available"; }
function onSelectionChange(rows: ExchangeCodeRecord[]): void {
  const selected = new Set(selectedCodeIds.value);
  codes.value.forEach((row) => selected.delete(row.id));
  rows.filter((row) => row.status === "available").forEach((row) => selected.add(row.id));
  if (selected.size > 500) { ElMessage.warning("最多只能批量选择 500 条"); return; }
  selectedCodeIds.value = [...selected];
}
async function copyCode(code: string): Promise<void> {
  try { await navigator.clipboard.writeText(code); ElMessage.success("兑换码已复制"); } catch { ElMessage.warning("复制失败，请手动复制"); }
}
function openBatchEdit(): void {
  Object.assign(batchForm, { edit_redeem_url: false, edit_batch_name: false, edit_expires_at: false, redeem_url: "", batch_name: "", expires_at: "" });
  batchEditVisible.value = true;
}
async function submitBatchEdit(): Promise<void> {
  if (!batchForm.edit_redeem_url && !batchForm.edit_batch_name && !batchForm.edit_expires_at) { ElMessage.warning("请选择至少一个要修改的字段"); return; }
  const payload: { ids: string[]; redeem_url?: string; batch_name?: string; expires_at?: string } = { ids: [...selectedCodeIds.value] };
  if (batchForm.edit_redeem_url) payload.redeem_url = batchForm.redeem_url.trim();
  if (batchForm.edit_batch_name) payload.batch_name = batchForm.batch_name.trim();
  if (batchForm.edit_expires_at) payload.expires_at = batchForm.expires_at.trim();
  batchEditing.value = true;
  try { const result = await batchUpdateExchangeCodes(payload); ElMessage.success(`已修改 ${result.updated} 条，跳过 ${result.skipped} 条`); batchEditVisible.value = false; selectedCodeIds.value = []; await loadInventory(); await loadCodes(); } catch { ElMessage.error("批量编辑失败，请检查字段内容"); } finally { batchEditing.value = false; }
}
async function deleteSelectedCodes(): Promise<void> {
  if (!selectedCodeIds.value.length) return;
  try { await ElMessageBox.confirm(`确认删除选中的 ${selectedCodeIds.value.length} 条可用兑换码？已分配和已使用记录不会删除。`, "批量删除库存", { type: "warning", confirmButtonText: "确认删除", cancelButtonText: "取消" }); } catch { return; }
  try { const result = await batchDeleteExchangeCodes(selectedCodeIds.value); ElMessage.success(`已删除 ${result.deleted} 条，跳过 ${result.skipped} 条`); selectedCodeIds.value = []; await loadInventory(); await loadCodes(); } catch { ElMessage.error("批量删除失败"); }
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
.inventory-toolbar, .drawer-toolbar, .inventory-summary-actions, .selection-toolbar { display: flex; gap: 10px; align-items: center; }
.drawer-toolbar { margin-bottom: 12px; }
.drawer-total { margin-left: auto; color: var(--app-muted); font-size: 13px; }
.inventory-summary-actions { justify-content: space-between; margin-top: 12px; }
.selection-toolbar { margin-bottom: 12px; flex-wrap: wrap; }
.inventory-table { width: 100%; }
.code-cell { display: inline-block; max-width: calc(100% - 34px); overflow: hidden; text-overflow: ellipsis; vertical-align: middle; white-space: nowrap; }
.drawer-pagination { display: flex; justify-content: flex-end; margin-top: 16px; overflow-x: auto; }
.el-dialog .el-checkbox { display: block; margin: 12px 0 6px; }
.el-dialog .el-input { margin-bottom: 8px; }
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
