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

    <PanelSection title="每日额度抽奖" description="每个用户每天免费抽 3 次；免费次数用完后，可按配置继续消耗积分抽奖。奖品从独立抽奖库存中发放。">
      <template #actions>
        <ChatSelect v-model="selectedChatId" />
        <el-button :icon="Refresh" :loading="dailyLoading" @click="loadDailyLottery">刷新配置</el-button>
        <el-button type="primary" :loading="dailySaving" @click="saveDailyLottery">保存每日抽奖</el-button>
      </template>
      <el-alert class="daily-alert" type="info" :closable="false" title="权重按千分比计算：总和 1000 表示每次必中奖；低于 1000 的部分表示未中奖概率。第 3 次可开启保底。" />
      <el-form label-position="top" class="daily-form">
        <div class="daily-switches">
          <div class="daily-switch-row"><span>开启每日额度抽奖</span><el-switch v-model="dailyForm.enabled" /></div>
          <div class="daily-switch-row"><span>第 3 次未中奖时发放奖品</span><el-switch v-model="dailyForm.guarantee_on_last" /></div>
          <div class="daily-switch-row"><span>开启积分抽奖</span><el-switch v-model="dailyForm.paid_enabled" /></div>
          <div class="daily-fixed-row"><span>每日次数</span><strong>3 次（固定）</strong></div>
        </div>
        <el-form-item label="积分抽奖每次消耗" class="daily-cost">
          <el-input-number v-model="dailyForm.cost_points" :min="1" :max="999999" :disabled="!dailyForm.paid_enabled" />
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

      <div class="daily-prize-header inventory-header">
        <div><strong>抽奖库存</strong><span>与普通积分兑换库存隔离；明细在分页抽屉中管理。</span></div>
        <div class="inventory-actions"><el-button :icon="Refresh" :loading="inventoryLoading" @click="loadDailyInventory">刷新摘要</el-button><el-button type="primary" @click="openDailyInventoryDrawer">管理库存</el-button></div>
      </div>
      <div class="inventory-summary-grid">
        <button v-for="item in dailyCodeSummaries" :key="item.amount" type="button" class="inventory-summary" :class="{ active: dailyInventory.amount === item.amount }" @click="selectDailyAmount(item.amount)">
          <strong>{{ item.amount }} 额度</strong>
          <span>可用 {{ item.available }} · 已分配 {{ item.assigned }} · 已使用 {{ item.used }}</span>
        </button>
        <button v-if="dailyInventory.amount !== undefined" type="button" class="clear-amount" @click="selectDailyAmount(undefined)">查看全部</button>
      </div>
      <el-form label-position="top" class="daily-inventory-form">
        <el-row :gutter="12">
          <el-col :xs="24" :md="6"><el-form-item label="额度"><el-input-number v-model="dailyInventory.amountInput" :min="1" :max="999999999" class="wide-control" /></el-form-item></el-col>
          <el-col :xs="24" :md="6"><el-form-item label="批次名称"><el-input v-model="dailyInventory.batchName" placeholder="例如：10额度首批" /></el-form-item></el-col>
          <el-col :xs="24" :md="12"><el-form-item label="兑换地址（可选）"><el-input v-model="dailyInventory.redeemURL" placeholder="https://example.com/redeem" /></el-form-item></el-col>
        </el-row>
        <el-form-item label="兑换码（每行一个）"><el-input v-model="dailyInventory.codes" type="textarea" :rows="5" placeholder="CODE-001\nCODE-002\nCODE-003" /></el-form-item>
        <el-button type="primary" :loading="inventoryImporting" @click="importDailyInventory">导入抽奖库存</el-button>
      </el-form>

      <div class="daily-prize-header reset-header">
        <div><strong>测试工具</strong><span>仅重置指定用户当天的抽奖次数；不会回收已经发出的兑换码，也不会返还已扣积分。</span></div>
      </div>
      <el-form label-position="top" class="daily-reset-form" @submit.prevent>
        <el-form-item label="Telegram 用户 ID">
          <el-input-number v-model="dailyReset.userId" :min="1" :max="999999999999999" :precision="0" class="wide-control" placeholder="例如：123456789" />
        </el-form-item>
        <el-button type="warning" :loading="dailyResetting" @click="resetDailyAttempts">重置今日抽奖次数</el-button>
      </el-form>
    </PanelSection>

    <el-drawer v-model="dailyInventoryDrawer" title="每日抽奖库存" size="min(920px, 95vw)" destroy-on-close @closed="selectedDailyCodeIds = []">
      <div class="drawer-toolbar">
        <el-select v-model="dailyInventory.status" clearable placeholder="全部状态" @change="reloadDailyInventoryDrawer">
          <el-option label="可用" value="available" /><el-option label="已分配" value="assigned" /><el-option label="已使用" value="used" />
        </el-select>
        <el-button :icon="Refresh" :loading="inventoryLoading" @click="loadDailyInventoryPage">刷新</el-button>
        <span class="drawer-total">共 {{ dailyInventoryTotal }} 条</span>
      </div>
      <div v-if="selectedDailyCodeIds.length" class="selection-toolbar">
        <strong>已选 {{ selectedDailyCodeIds.length }} 条</strong>
        <el-button type="primary" @click="openDailyBatchEdit">批量编辑</el-button>
        <el-button type="danger" @click="deleteSelectedDailyCodes">批量删除</el-button>
        <span class="field-hint">仅可用兑换码可编辑或删除</span>
      </div>
      <el-table :data="dailyCodes" row-key="id" stripe size="small" class="inventory-table" @selection-change="onDailySelectionChange">
        <el-table-column type="selection" width="48" :selectable="isDailyCodeSelectable" reserve-selection />
        <el-table-column prop="code" label="兑换码" min-width="190" show-overflow-tooltip>
          <template #default="{ row }"><span class="code-cell">{{ row.code }}</span><el-button text aria-label="复制兑换码" @click="copyDailyCode(row.code)">复制</el-button></template>
        </el-table-column>
        <el-table-column prop="amount" label="额度" width="80" />
        <el-table-column prop="batch_name" label="批次" min-width="130" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="90" />
        <el-table-column prop="redeem_url" label="兑换地址" min-width="180" show-overflow-tooltip />
      </el-table>
      <div class="drawer-pagination"><el-pagination v-model:current-page="dailyInventoryPage" v-model:page-size="dailyInventoryPageSize" :page-sizes="[20, 50, 100]" layout="total, sizes, prev, pager, next" :total="dailyInventoryTotal" @current-change="loadDailyInventoryPage" @size-change="handleDailyInventorySizeChange" /></div>
    </el-drawer>

    <el-dialog v-model="dailyBatchEditVisible" title="批量编辑抽奖库存" width="500px">
      <p class="field-hint">已选择 {{ selectedDailyCodeIds.length }} 条，仅修改勾选的字段；空值会清空对应字段。</p>
      <el-form label-position="top">
        <el-checkbox v-model="dailyBatchForm.edit_redeem_url">修改兑换地址</el-checkbox>
        <el-input v-model="dailyBatchForm.redeem_url" :disabled="!dailyBatchForm.edit_redeem_url" placeholder="留空可清除，或填写 https://..." />
        <el-checkbox v-model="dailyBatchForm.edit_batch_name">修改批次名称</el-checkbox>
        <el-input v-model="dailyBatchForm.batch_name" :disabled="!dailyBatchForm.edit_batch_name" maxlength="128" placeholder="留空可清除" />
        <el-checkbox v-model="dailyBatchForm.edit_expires_at">修改有效期</el-checkbox>
        <el-input v-model="dailyBatchForm.expires_at" :disabled="!dailyBatchForm.edit_expires_at" placeholder="RFC3339 时间；留空清除" />
      </el-form>
      <template #footer><el-button @click="dailyBatchEditVisible = false">取消</el-button><el-button type="primary" :loading="dailyBatchEditing" @click="submitDailyBatchEdit">保存修改</el-button></template>
    </el-dialog>

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
import { CopyDocument, Delete, Edit, MoreFilled, Plus, Refresh } from "@element-plus/icons-vue";
import ChatSelect from "@/components/ChatSelect.vue";
import PageHeader from "@/components/PageHeader.vue";
import PanelSection from "@/components/PanelSection.vue";
import { cancelLottery, createLottery, fetchLotteries, fetchLotteryEntries, fetchLotteryWinners } from "@/api/lottery";
import { batchDeleteDailyLotteryCodes, batchUpdateDailyLotteryCodes, fetchDailyLotteryCodeSummary, fetchDailyLotteryCodes, fetchDailyLotteryConfig, fetchDailyLotteryPrizes, importDailyLotteryCodes, resetDailyLotteryAttempts, updateDailyLottery, type DailyLotteryCode, type DailyLotteryCodeSummary, type DailyLotteryPrize } from "@/api/dailyLottery";
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
const dailyCodes = ref<DailyLotteryCode[]>([]);
const dailyCodeSummaries = ref<DailyLotteryCodeSummary[]>([]);
const inventoryLoading = ref(false);
const inventoryImporting = ref(false);
const dailyInventoryDrawer = ref(false);
const dailyInventoryPage = ref(1);
const dailyInventoryPageSize = ref(20);
const dailyInventoryTotal = ref(0);
const selectedDailyCodeIds = ref<string[]>([]);
const dailyBatchEditVisible = ref(false);
const dailyBatchEditing = ref(false);
const dailyBatchForm = reactive({ edit_redeem_url: false, edit_batch_name: false, edit_expires_at: false, redeem_url: "", batch_name: "", expires_at: "" });
const dailyResetting = ref(false);
const dailyInventory = reactive({ amount: undefined as number | undefined, amountInput: 10, batchName: "", redeemURL: "", codes: "", status: "" });
const dailyReset = reactive({ userId: undefined as number | undefined });
const dailyForm = reactive({ enabled: false, daily_attempts: 3, cost_points: 1, paid_enabled: false, guarantee_on_last: true });
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
    dailyCodes.value = [];
    dailyCodeSummaries.value = [];
    dailyInventoryTotal.value = 0;
    Object.assign(dailyForm, { enabled: false, daily_attempts: 3, cost_points: 1, paid_enabled: false, guarantee_on_last: true });
    return;
  }
  dailyLoading.value = true;
  try {
    const [config, prizes] = await Promise.all([fetchDailyLotteryConfig(selectedChatId.value), fetchDailyLotteryPrizes(selectedChatId.value)]);
    Object.assign(dailyForm, config);
    dailyPrizes.value = prizes.items;
    await loadDailyInventory();
  } catch (error) {
    ElMessage.error(errorMessage(error));
  } finally {
    dailyLoading.value = false;
  }
}

async function loadDailyInventory(): Promise<void> {
  if (!selectedChatId.value) { dailyCodes.value = []; dailyCodeSummaries.value = []; return; }
  inventoryLoading.value = true;
  try {
    const summaries = await fetchDailyLotteryCodeSummary(selectedChatId.value);
    dailyCodeSummaries.value = summaries.items;
  } catch (error) { ElMessage.error(errorMessage(error)); }
  finally { inventoryLoading.value = false; }
}

async function loadDailyInventoryPage(): Promise<void> {
  if (!selectedChatId.value || !dailyInventoryDrawer.value) return;
  inventoryLoading.value = true;
  try {
    const result = await fetchDailyLotteryCodes(selectedChatId.value, dailyInventory.amount, dailyInventory.status, dailyInventoryPage.value, dailyInventoryPageSize.value);
    dailyCodes.value = result.items;
    dailyInventoryTotal.value = result.total;
  } catch (error) { ElMessage.error(errorMessage(error)); }
  finally { inventoryLoading.value = false; }
}

function openDailyInventoryDrawer(): void { dailyInventoryDrawer.value = true; dailyInventoryPage.value = 1; void loadDailyInventoryPage(); }
function reloadDailyInventoryDrawer(): void { dailyInventoryPage.value = 1; selectedDailyCodeIds.value = []; void loadDailyInventoryPage(); }
function handleDailyInventorySizeChange(size: number): void { dailyInventoryPageSize.value = size; dailyInventoryPage.value = 1; void loadDailyInventoryPage(); }
function isDailyCodeSelectable(row: DailyLotteryCode): boolean { return row.status === "available"; }
function onDailySelectionChange(rows: DailyLotteryCode[]): void {
  const selected = new Set(selectedDailyCodeIds.value);
  dailyCodes.value.forEach((row) => selected.delete(row.id));
  rows.filter((row) => row.status === "available").forEach((row) => selected.add(row.id));
  if (selected.size > 500) { ElMessage.warning("最多只能批量选择 500 条"); return; }
  selectedDailyCodeIds.value = [...selected];
}
async function copyDailyCode(code: string): Promise<void> {
  try { await navigator.clipboard.writeText(code); ElMessage.success("兑换码已复制"); } catch { ElMessage.warning("复制失败，请手动复制"); }
}
function openDailyBatchEdit(): void {
  Object.assign(dailyBatchForm, { edit_redeem_url: false, edit_batch_name: false, edit_expires_at: false, redeem_url: "", batch_name: "", expires_at: "" });
  dailyBatchEditVisible.value = true;
}
async function submitDailyBatchEdit(): Promise<void> {
  if (!dailyBatchForm.edit_redeem_url && !dailyBatchForm.edit_batch_name && !dailyBatchForm.edit_expires_at) { ElMessage.warning("请选择至少一个要修改的字段"); return; }
  const chatID = parseNumericId(selectedChatId.value);
  if (chatID === undefined) { ElMessage.warning("请选择有效群组"); return; }
  const payload: { chat_id: ChatID; ids: string[]; redeem_url?: string; batch_name?: string; expires_at?: string } = { chat_id: chatID, ids: [...selectedDailyCodeIds.value] };
  if (dailyBatchForm.edit_redeem_url) payload.redeem_url = dailyBatchForm.redeem_url.trim();
  if (dailyBatchForm.edit_batch_name) payload.batch_name = dailyBatchForm.batch_name.trim();
  if (dailyBatchForm.edit_expires_at) payload.expires_at = dailyBatchForm.expires_at.trim();
  dailyBatchEditing.value = true;
  try { const result = await batchUpdateDailyLotteryCodes(payload); ElMessage.success(`已修改 ${result.updated} 条，跳过 ${result.skipped} 条`); dailyBatchEditVisible.value = false; selectedDailyCodeIds.value = []; await loadDailyInventory(); await loadDailyInventoryPage(); } catch (error) { ElMessage.error(errorMessage(error)); } finally { dailyBatchEditing.value = false; }
}
async function deleteSelectedDailyCodes(): Promise<void> {
  const chatID = parseNumericId(selectedChatId.value);
  if (chatID === undefined || !selectedDailyCodeIds.value.length) return;
  try { await ElMessageBox.confirm(`确认删除选中的 ${selectedDailyCodeIds.value.length} 条可用兑换码？已分配和已使用记录不会删除。`, "批量删除抽奖库存", { type: "warning", confirmButtonText: "确认删除", cancelButtonText: "取消" }); } catch { return; }
  try { const result = await batchDeleteDailyLotteryCodes({ chat_id: chatID, ids: selectedDailyCodeIds.value }); ElMessage.success(`已删除 ${result.deleted} 条，跳过 ${result.skipped} 条`); selectedDailyCodeIds.value = []; await loadDailyInventory(); await loadDailyInventoryPage(); } catch (error) { ElMessage.error(errorMessage(error)); }
}

function selectDailyAmount(amount?: number): void {
  dailyInventory.amount = amount;
  if (amount !== undefined) dailyInventory.amountInput = amount;
  dailyInventoryPage.value = 1;
  void loadDailyInventory();
  if (dailyInventoryDrawer.value) void loadDailyInventoryPage();
}

async function importDailyInventory(): Promise<void> {
  const chatID = parseNumericId(selectedChatId.value);
  if (chatID === undefined || !Number.isSafeInteger(chatID)) { ElMessage.warning("请选择有效群组"); return; }
  if (!Number.isSafeInteger(Number(dailyInventory.amountInput)) || dailyInventory.amountInput <= 0) { ElMessage.warning("请输入有效额度"); return; }
  const codes = dailyInventory.codes.split(/\r?\n/).map((item) => item.trim()).filter(Boolean);
  if (!codes.length) { ElMessage.warning("请每行填写一个兑换码"); return; }
  inventoryImporting.value = true;
  try {
    const result = await importDailyLotteryCodes({ chat_id: chatID, amount: Number(dailyInventory.amountInput), batch_name: dailyInventory.batchName.trim(), redeem_url: dailyInventory.redeemURL.trim(), codes });
    ElMessage.success(`导入 ${result.imported} 个，跳过 ${result.skipped} 个`);
    dailyInventory.codes = "";
    await loadDailyInventory();
    await loadDailyLottery();
  } catch (error) { ElMessage.error(errorMessage(error)); }
  finally { inventoryImporting.value = false; }
}

async function saveDailyLottery(): Promise<void> {
  const chatID = parseNumericId(selectedChatId.value);
  if (chatID === undefined || !Number.isSafeInteger(chatID)) {
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
	if (dailyPrizes.value.length > 0 && (dailyWeightTotal.value <= 0 || dailyWeightTotal.value > 1000)) {
		ElMessage.warning("启用奖池权重合计不能超过 1000");
		return;
	}
  if (dailyForm.paid_enabled && (!Number.isSafeInteger(Number(dailyForm.cost_points)) || Number(dailyForm.cost_points) <= 0)) {
    ElMessage.warning("开启积分抽奖后，请设置大于 0 的积分消耗");
    return;
  }
  dailySaving.value = true;
  try {
    await updateDailyLottery({
      chat_id: chatID,
      enabled: dailyForm.enabled,
      cost_points: dailyForm.cost_points,
      paid_enabled: dailyForm.paid_enabled,
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

async function resetDailyAttempts(): Promise<void> {
  const chatID = parseNumericId(selectedChatId.value);
  const userID = parseNumericId(dailyReset.userId);
  if (chatID === undefined || !Number.isSafeInteger(chatID)) { ElMessage.warning("请先选择群组"); return; }
  if (userID === undefined || !Number.isSafeInteger(userID) || userID <= 0) { ElMessage.warning("请输入有效的 Telegram 用户 ID"); return; }
  try {
    await ElMessageBox.confirm("将清空该用户今天的每日抽奖记录，已发出的兑换码和已扣积分不会撤销。确认继续？", "重置今日抽奖次数", {
      type: "warning",
      confirmButtonText: "确认重置",
      cancelButtonText: "取消",
    });
  } catch {
    return;
  }
  dailyResetting.value = true;
  try {
    await resetDailyLotteryAttempts({ chat_id: chatID, user_id: userID });
    ElMessage.success("该用户今日抽奖次数已重置");
  } catch (error) {
    ElMessage.error(errorMessage(error));
  } finally {
    dailyResetting.value = false;
  }
}

watch(selectedChatId, () => {
  dailyInventoryPage.value = 1;
  selectedDailyCodeIds.value = [];
  void loadDailyLottery();
  if (dailyInventoryDrawer.value) void loadDailyInventoryPage();
});
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
.inventory-actions, .drawer-toolbar, .selection-toolbar { display: flex; align-items: center; gap: 10px; }
.drawer-toolbar { margin-bottom: 12px; }
.drawer-total { margin-left: auto; color: var(--app-muted); font-size: 13px; }
.selection-toolbar { flex-wrap: wrap; margin-bottom: 12px; }
.inventory-table { width: 100%; }
.code-cell { display: inline-block; max-width: calc(100% - 48px); overflow: hidden; text-overflow: ellipsis; vertical-align: middle; white-space: nowrap; }
.drawer-pagination { display: flex; justify-content: flex-end; margin-top: 16px; overflow-x: auto; }
.el-dialog .el-checkbox { display: block; margin: 12px 0 6px; }
.el-dialog .el-input { margin-bottom: 8px; }
.daily-total {
  margin-top: 10px;
}
.inventory-header {
  margin-top: 24px;
}
.reset-header {
  margin-top: 24px;
}
.daily-reset-form {
  display: flex;
  align-items: flex-end;
  gap: 12px;
  max-width: 360px;
}
.daily-reset-form .el-form-item {
  flex: 1;
  margin-bottom: 0;
}
.daily-inventory-form {
  margin-top: 12px;
}
.daily-inventory-table {
  margin-top: 16px;
}
.inventory-summary-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin: 12px 0;
}
.inventory-summary, .clear-amount {
  border: 1px solid var(--app-border);
  border-radius: 8px;
  background: var(--app-surface-2);
  padding: 8px 10px;
  color: var(--app-text);
  cursor: pointer;
  text-align: left;
}
.inventory-summary { display: flex; min-width: 140px; flex-direction: column; gap: 3px; }
.inventory-summary.active, .inventory-summary:hover, .clear-amount:hover { border-color: var(--app-primary); }
.inventory-summary span { color: var(--app-muted); font-size: 12px; }
@media (max-width: 900px) {
  .daily-form { align-items: stretch; flex-direction: column; }
  .daily-switches { grid-template-columns: 1fr; }
  .daily-cost { min-width: 0; }
  .daily-reset-form { align-items: stretch; flex-direction: column; }
}
</style>
