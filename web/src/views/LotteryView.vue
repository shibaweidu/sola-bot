<template>
  <div class="page-stack">
    <PageHeader eyebrow="活动运营" title="活动抽奖" description="管理群内抽奖活动、参与方式和开奖结果。">
      <template #meta>
        <span class="header-meta">页面时间按北京时间（东八区）显示</span>
      </template>
      <template #actions>
        <el-button :icon="Refresh" :loading="loading" @click="loadLotteries">刷新</el-button>
        <el-button type="primary" :icon="Plus" @click="openCreate">创建抽奖</el-button>
      </template>
    </PageHeader>

    <PanelSection title="每日额度抽奖" description="每个用户在当前群组每天可抽 3 次，奖品从现有额度兑换码库存中发放。">
      <template #actions>
        <ChatSelect v-model="selectedChatId" />
        <el-button :icon="Refresh" :loading="dailyLoading" @click="loadDailyLottery">刷新配置</el-button>
        <el-button type="primary" :loading="dailySaving" @click="saveDailyLottery">保存每日抽奖</el-button>
      </template>
      <el-alert class="daily-alert" type="info" :closable="false" title="权重按千分比计算：总和 1000 表示每次必中奖；低于 1000 的部分表示未中奖概率。第 3 次可开启保底。" />
      <el-form label-position="top" class="daily-form">
        <div class="daily-switches">
          <div class="daily-switch-row"><span>开启每日额度抽奖</span><el-switch v-model="dailyForm.enabled" /></div>
          <div class="daily-switch-row"><span>第 3 次未中奖时强制发码</span><el-switch v-model="dailyForm.guarantee_on_last" /></div>
          <div class="daily-fixed-row"><span>每日次数</span><strong>3 次（固定）</strong></div>
        </div>
        <el-form-item label="每次抽奖消耗积分（0 = 免费）" class="daily-cost">
          <el-input-number v-model="dailyForm.cost_points" :min="0" :max="999999" />
        </el-form-item>
      </el-form>
      <div class="daily-prize-header">
        <div><strong>额度奖池</strong><span>同一额度只能配置一次；无库存额度会自动跳过。</span></div>
        <el-button type="primary" plain @click="addDailyPrize">添加额度</el-button>
      </div>
      <el-table :data="dailyPrizes" stripe size="small" empty-text="请添加至少一个额度奖池">
        <el-table-column label="额度" min-width="150">
          <template #default="{ row }"><el-input-number v-model="row.amount" :min="1" :max="999999999" /></template>
        </el-table-column>
        <el-table-column label="权重（/1000）" min-width="170">
          <template #default="{ row }"><el-input-number v-model="row.weight" :min="1" :max="1000" /></template>
        </el-table-column>
        <el-table-column label="预计比例" width="110">
          <template #default="{ row }">{{ dailyPrizePercent(row.weight) }}</template>
        </el-table-column>
        <el-table-column prop="available_code" label="可用兑换码" width="120" />
        <el-table-column label="启用" width="80"><template #default="{ row }"><el-switch v-model="row.enabled" /></template></el-table-column>
        <el-table-column label="操作" width="90"><template #default="{ $index }"><el-button type="danger" link @click="removeDailyPrize($index)">删除</el-button></template></el-table-column>
      </el-table>
      <div class="daily-total">启用奖池权重合计：{{ dailyWeightTotal }}/1000 · {{ dailyPrizePercent(dailyWeightTotal) }}</div>
    </PanelSection>

    <div class="summary-grid">
      <div class="summary-card">
        <div class="summary-label">当前活动</div>
        <div class="summary-value">{{ filteredLotteries.length }}</div>
        <div class="summary-meta">匹配当前筛选条件</div>
      </div>
      <div class="summary-card">
        <div class="summary-label">进行中</div>
        <div class="summary-value">{{ statusCounts.active }}</div>
        <div class="summary-meta">待开奖活动</div>
      </div>
      <div class="summary-card">
        <div class="summary-label">已开奖</div>
        <div class="summary-value">{{ statusCounts.ended }}</div>
        <div class="summary-meta">已完成结果发放</div>
      </div>
      <div class="summary-card">
        <div class="summary-label">参与次数</div>
        <div class="summary-value">{{ totalEntries }}</div>
        <div class="summary-meta">当前列表累计参与</div>
      </div>
    </div>

    <PanelSection title="抽奖列表" description="支持按群组、状态和参与方式查看活动进度。">
      <template #actions>
        <div class="panel-toolbar">
          <div class="control-cluster filters">
            <div class="filter-control filter-control-wide">
              <ChatSelect v-model="selectedChatId" />
            </div>
            <el-select v-model="statusFilter" class="filter-control" clearable placeholder="状态">
              <el-option label="进行中" value="active" />
              <el-option label="已开奖" value="ended" />
              <el-option label="已取消" value="cancelled" />
            </el-select>
            <el-select v-model="joinTypeFilter" class="filter-control" clearable placeholder="参与方式">
              <el-option label="按钮参与" value="button" />
              <el-option label="口令参与" value="keyword" />
              <el-option label="按钮 + 口令" value="both" />
            </el-select>
          </div>
          <div class="filter-summary">
            <span>群组 {{ selectedChatId || '全部' }}</span>
            <span>状态 {{ statusFilter || '全部' }}</span>
            <span>参与方式 {{ joinTypeFilter || '全部' }}</span>
          </div>
        </div>
      </template>

      <div class="table-wrap">
        <el-table class="table-compact" :data="filteredLotteries" size="small" stripe>
        <el-table-column prop="title" label="标题" min-width="180" />
        <el-table-column prop="prize" label="奖品" min-width="140" />
        <el-table-column prop="chat_id" label="目标群组" min-width="140" />
        <el-table-column prop="cost_points" label="积分成本" width="100" />
        <el-table-column label="参与方式" width="130">
          <template #default="{ row }">
            <el-tag :type="joinTypeTag(row.join_type)" effect="plain">{{ joinTypeLabel(row.join_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="参与人数" width="110">
          <template #default="{ row }">{{ row.entry_count ?? row.participants ?? 0 }}</template>
        </el-table-column>
        <el-table-column prop="winner_count" label="中奖数" width="90" />
        <el-table-column label="开奖时间（北京时间）" min-width="180">
          <template #default="{ row }">{{ formatDateTime(row.end_at) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="statusTag(row.status)" effect="dark">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="190" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="showEntries(row)">参与者</el-button>
            <el-dropdown @command="handleLotteryCommand($event, row)">
              <el-button size="small">
                更多
                <el-icon class="el-icon--right"><MoreFilled /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="winners">中奖名单</el-dropdown-item>
                  <el-dropdown-item command="cancel" :disabled="row.status !== 'active'">取消活动</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
        </el-table-column>
      </el-table>
      </div>
    </PanelSection>

    <el-dialog v-model="dialogVisible" title="创建抽奖" width="560px">
      <el-form label-position="top">
        <el-row :gutter="12">
          <el-col :xs="24" :md="12">
            <el-form-item label="目标群组">
              <ChatSelect v-model="form.chat_id" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :md="12">
            <el-form-item label="开奖时间（北京时间）">
              <el-date-picker
                v-model="form.end_at"
                class="wide-control"
                type="datetime"
                format="YYYY-MM-DD HH:mm"
                value-format="YYYY-MM-DD HH:mm:ss"
                placeholder="选择开奖时间（北京时间）"
              />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="标题">
          <el-input v-model="form.title" />
        </el-form-item>
        <el-form-item label="奖品">
          <el-input v-model="form.prize" />
        </el-form-item>
        <el-row :gutter="12">
          <el-col :xs="24" :md="12">
            <el-form-item label="参与方式">
              <el-segmented v-model="form.join_type" :options="joinTypeOptions" class="wide-control" />
            </el-form-item>
          </el-col>
          <el-col v-if="requiresJoinKeyword" :xs="24" :md="12">
            <el-form-item label="参与口令">
              <el-input v-model="form.join_keyword" maxlength="64" show-word-limit placeholder="例如 888" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="12">
          <el-col :xs="24" :md="8">
            <el-form-item label="参与成本">
              <el-input-number v-model="form.cost_points" class="wide-control" :min="0" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :md="8">
            <el-form-item label="人数上限">
              <el-input-number v-model="form.max_participants" class="wide-control" :min="0" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :md="8">
            <el-form-item label="中奖人数">
              <el-input-number v-model="form.winner_count" class="wide-control" :min="1" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submitLottery">创建</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="entriesVisible" :title="entriesTitle" width="560px">
      <el-table class="table-compact" :data="entries" size="small" stripe>
        <el-table-column prop="user_id" label="用户 ID" min-width="150" />
        <el-table-column label="用户名" min-width="130">
          <template #default="{ row }">{{ row.username ? `@${row.username}` : '-' }}</template>
        </el-table-column>
        <el-table-column prop="joined_at" label="参与时间" min-width="170" />
        <el-table-column label="中奖" width="90">
          <template #default="{ row }">
            <el-tag :type="row.is_winner ? 'success' : 'info'" effect="dark">{{ row.is_winner ? '是' : '否' }}</el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { MoreFilled, Plus, Refresh } from "@element-plus/icons-vue";
import ChatSelect from "@/components/ChatSelect.vue";
import PageHeader from "@/components/PageHeader.vue";
import PanelSection from "@/components/PanelSection.vue";
import { cancelLottery, createLottery, fetchLotteries, fetchLotteryEntries, fetchLotteryWinners } from "@/api/lottery";
import { fetchDailyLotteryConfig, fetchDailyLotteryPrizes, updateDailyLottery, type DailyLotteryPrize } from "@/api/dailyLottery";
import type { ChatID, LotteryEntryRecord, LotteryPayload, LotteryRecord } from "@/types/api";
import { parseChinaLocalDateTimeToISO } from "@/utils/datetime";
import { parseNumericId, formatDateTime, errorMessage } from "@/utils/helpers";

const loading = ref(false);
const saving = ref(false);
const dialogVisible = ref(false);
const entriesVisible = ref(false);
const lotteries = ref<LotteryRecord[]>([]);
const entries = ref<LotteryEntryRecord[]>([]);
const entriesSeq = ref(0);
const entriesTitle = ref("参与者");
const cancellingId = ref<ChatID>();
const selectedChatId = ref<ChatID | "">("");
const statusFilter = ref<LotteryRecord["status"] | "">("");
const joinTypeFilter = ref<LotteryRecord["join_type"] | "">("");
const dailyLoading = ref(false);
const dailySaving = ref(false);
const dailyPrizes = ref<DailyLotteryPrize[]>([]);
const dailyForm = reactive({ enabled: false, daily_attempts: 3, cost_points: 0, guarantee_on_last: true });
const form = reactive<LotteryPayload>({
  chat_id: "",
  title: "",
  prize: "",
  cost_points: 0,
  max_participants: 0,
  winner_count: 1,
  end_at: "",
  created_by: "",
  join_type: "button",
  join_keyword: "",
});
const joinTypeOptions = [
  { label: "按钮", value: "button" },
  { label: "口令", value: "keyword" },
  { label: "按钮+口令", value: "both" },
];
const requiresJoinKeyword = computed(() => form.join_type === "keyword" || form.join_type === "both");

const filteredLotteries = computed(() => {
  return lotteries.value.filter((item) => {
    const matchesChat = selectedChatId.value ? String(item.chat_id) === String(selectedChatId.value) : true;
    const matchesStatus = statusFilter.value ? item.status === statusFilter.value : true;
    const matchesJoinType = joinTypeFilter.value ? item.join_type === joinTypeFilter.value : true;
    return matchesChat && matchesStatus && matchesJoinType;
  });
});

const statusCounts = computed(() => {
  return filteredLotteries.value.reduce(
    (acc, item) => {
      acc[item.status] += 1;
      return acc;
    },
    { active: 0, ended: 0, cancelled: 0 } as Record<LotteryRecord["status"], number>,
  );
});

const totalEntries = computed(() => filteredLotteries.value.reduce((sum, item) => sum + Number(item.entry_count ?? item.participants ?? 0), 0));
const dailyWeightTotal = computed(() => dailyPrizes.value.filter((item) => item.enabled).reduce((sum, item) => sum + Number(item.weight || 0), 0));

function optionalText(value?: string | null): string | undefined {
  const text = value?.trim();
  return text || undefined;
}

function toRFC3339(value?: string | null): string | undefined {
  return parseChinaLocalDateTimeToISO(value);
}

function buildPayload(): LotteryPayload | undefined {
  const chatId = parseNumericId(form.chat_id);
  if (!chatId) {
    ElMessage.warning("请输入有效的群组 ID");
    return undefined;
  }
  if (!form.title.trim()) {
    ElMessage.warning("请输入抽奖标题");
    return undefined;
  }
  const joinType = form.join_type || "button";
  const joinKeyword = optionalText(form.join_keyword);
  if ((joinType === "keyword" || joinType === "both") && !joinKeyword) {
    ElMessage.warning("请输入参与口令");
    return undefined;
  }
  return {
    chat_id: chatId,
    title: form.title.trim(),
    prize: form.prize.trim(),
    cost_points: form.cost_points,
    max_participants: form.max_participants,
    winner_count: form.winner_count,
    end_at: toRFC3339(form.end_at),
    created_by: parseNumericId(form.created_by) ?? 0,
    join_type: joinType,
    join_keyword: joinType === "button" ? undefined : joinKeyword,
  };
}

function openCreate(): void {
  Object.assign(form, {
    chat_id: selectedChatId.value ? String(selectedChatId.value) : "",
    title: "",
    prize: "",
    cost_points: 0,
    max_participants: 0,
    winner_count: 1,
    end_at: "",
    created_by: "",
    join_type: "button",
    join_keyword: "",
  });
  dialogVisible.value = true;
}

async function loadLotteries(): Promise<void> {
  loading.value = true;
  try {
    lotteries.value = await fetchLotteries();
  } catch (error) {
    lotteries.value = [];
    ElMessage.error(errorMessage(error));
  } finally {
    loading.value = false;
  }
}

async function submitLottery(): Promise<void> {
  const payload = buildPayload();
  if (!payload) return;
  saving.value = true;
  try {
    await createLottery(payload);
    ElMessage.success("抽奖已创建");
    dialogVisible.value = false;
    await loadLotteries();
  } catch (error) {
    ElMessage.error(errorMessage(error));
  } finally {
    saving.value = false;
  }
}

async function cancel(row: LotteryRecord): Promise<void> {
  try {
    await ElMessageBox.confirm(`确认取消抽奖「${row.title}」？`, "取消抽奖", {
      type: "warning",
      confirmButtonText: "取消抽奖",
      cancelButtonText: "返回",
    });
  } catch {
    return;
  }
  cancellingId.value = row.id;
  try {
    await cancelLottery(row.id);
    ElMessage.success("抽奖已取消");
    await loadLotteries();
  } catch (error) {
    ElMessage.error(errorMessage(error));
  } finally {
    cancellingId.value = undefined;
  }
}

async function showEntries(row: LotteryRecord): Promise<void> {
  const seq = ++entriesSeq.value;
  entriesTitle.value = `抽奖「${row.title || row.id}」参与者`;
  entriesVisible.value = true;
  try {
    const result = await fetchLotteryEntries(row.id);
    if (seq === entriesSeq.value) entries.value = result;
  } catch (error) {
    if (seq === entriesSeq.value) ElMessage.error(errorMessage(error));
  }
}

async function showWinners(row: LotteryRecord): Promise<void> {
  const seq = ++entriesSeq.value;
  entriesTitle.value = `抽奖「${row.title || row.id}」中奖名单`;
  entriesVisible.value = true;
  try {
    const result = await fetchLotteryWinners(row.id);
    if (seq === entriesSeq.value) entries.value = result;
  } catch (error) {
    if (seq === entriesSeq.value) ElMessage.error(errorMessage(error));
  }
}

function handleLotteryCommand(command: string, row: LotteryRecord): void {
  if (command === "winners") {
    void showWinners(row);
    return;
  }
  if (command === "cancel") {
    void cancel(row);
  }
}

function statusLabel(status: LotteryRecord["status"]): string {
  return { active: "进行中", ended: "已开奖", cancelled: "已取消" }[status];
}

function statusTag(status: LotteryRecord["status"]): "success" | "info" | "danger" {
  if (status === "active") return "success";
  if (status === "ended") return "info";
  return "danger";
}

function joinTypeLabel(joinType?: LotteryRecord["join_type"]): string {
  if (joinType === "keyword") return "口令参与";
  if (joinType === "both") return "按钮+口令";
  return "按钮参与";
}

function joinTypeTag(joinType?: LotteryRecord["join_type"]): "success" | "warning" | "primary" {
  if (joinType === "keyword") return "warning";
  if (joinType === "both") return "primary";
  return "success";
}

function dailyPrizePercent(weight: number): string {
  return `${((Math.max(0, Number(weight) || 0) / 1000) * 100).toFixed(1)}%`;
}

function addDailyPrize(): void {
  dailyPrizes.value.push({ chat_id: selectedChatId.value || "", amount: 10, weight: 100, enabled: true, available_code: 0 });
}

function removeDailyPrize(index: number): void {
  dailyPrizes.value.splice(index, 1);
}

async function loadDailyLottery(): Promise<void> {
  if (!selectedChatId.value) {
    dailyPrizes.value = [];
    Object.assign(dailyForm, { enabled: false, daily_attempts: 3, cost_points: 0, guarantee_on_last: true });
    return;
  }
  dailyLoading.value = true;
  try {
    const [config, prizes] = await Promise.all([fetchDailyLotteryConfig(selectedChatId.value), fetchDailyLotteryPrizes(selectedChatId.value)]);
    Object.assign(dailyForm, config);
    dailyPrizes.value = prizes.items;
  } catch (error) {
    ElMessage.error(errorMessage(error));
  } finally {
    dailyLoading.value = false;
  }
}

async function saveDailyLottery(): Promise<void> {
  if (!selectedChatId.value) {
    ElMessage.warning("请先选择群组");
    return;
  }
  const seen = new Set<number>();
  for (const prize of dailyPrizes.value) {
    const amount = Number(prize.amount);
    const weight = Number(prize.weight);
    if (!Number.isSafeInteger(amount) || amount <= 0) { ElMessage.warning("额度必须是正整数"); return; }
    if (seen.has(amount)) { ElMessage.warning("额度不能重复"); return; }
    seen.add(amount);
    if (!Number.isSafeInteger(weight) || weight < 1 || weight > 1000) { ElMessage.warning("权重必须在 1 到 1000 之间"); return; }
  }
  if (!dailyPrizes.value.length || dailyWeightTotal.value <= 0 || dailyWeightTotal.value > 1000) {
    ElMessage.warning("请配置至少一个奖池，启用权重合计不能超过 1000");
    return;
  }
  dailySaving.value = true;
  try {
    await updateDailyLottery({
      chat_id: selectedChatId.value,
      enabled: dailyForm.enabled,
      cost_points: dailyForm.cost_points,
      guarantee_on_last: dailyForm.guarantee_on_last,
      prizes: dailyPrizes.value.map(({ amount, weight, enabled }) => ({ amount, weight, enabled })),
    });
    ElMessage.success("每日额度抽奖配置已保存");
    await loadDailyLottery();
  } catch (error) {
    ElMessage.error(errorMessage(error));
  } finally {
    dailySaving.value = false;
  }
}

watch(selectedChatId, () => { void loadDailyLottery(); });
onMounted(() => { void loadLotteries(); void loadDailyLottery(); });
</script>

<style scoped>
.header-meta {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 5px 10px;
  border: 1px solid var(--app-border);
  border-radius: 999px;
  background: var(--app-tint-light);
}
.panel-toolbar {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: min(100%, 980px);
}

.filters :deep(.chat-select) {
  width: 100%;
}

.wide-control {
  width: 100%;
}
.daily-alert {
  margin-bottom: 16px;
}
.daily-form {
  display: flex;
  align-items: flex-end;
  gap: 16px;
  margin-bottom: 18px;
}
.daily-switches {
  display: grid;
  flex: 1;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}
.daily-switch-row, .daily-fixed-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 40px;
  border: 1px solid var(--app-border);
  border-radius: 8px;
  padding: 8px 10px;
}
.daily-fixed-row {
  color: var(--app-muted);
}
.daily-fixed-row strong {
  color: var(--app-text);
}
.daily-cost {
  min-width: 220px;
  margin-bottom: 0;
}
.daily-prize-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
}
.daily-prize-header > div {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.daily-prize-header span, .daily-total {
  color: var(--app-muted);
  font-size: 12px;
}
.daily-total {
  margin-top: 10px;
}
@media (max-width: 900px) {
  .daily-form { align-items: stretch; flex-direction: column; }
  .daily-switches { grid-template-columns: 1fr; }
  .daily-cost { min-width: 0; }
}
</style>
