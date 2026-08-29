<template>
  <div class="page">
    <PageHeader eyebrow="机器人运营" title="机器人菜单" description="配置私聊持久键盘。菜单按机器人全局生效，群组内仍使用原有按钮和命令。">
      <template #actions>
        <el-radio-group v-model="role" @change="loadMenu">
          <el-radio-button value="member">普通用户菜单</el-radio-button>
          <el-radio-button value="admin">管理员菜单</el-radio-button>
        </el-radio-group>
        <el-button :icon="Refresh" :loading="loading" @click="loadMenu">刷新</el-button>
        <el-button :icon="RefreshLeft" :loading="saving" @click="restoreDefaults">恢复默认</el-button>
        <el-button type="primary" :icon="Check" :loading="saving" @click="saveMenu">保存菜单</el-button>
      </template>
    </PageHeader>

    <el-row :gutter="16">
      <el-col :xs="24" :lg="16">
        <PanelSection :title="role === 'admin' ? '管理员按钮' : '普通用户按钮'" description="按钮文字可自定义，但动作仍受内置白名单和权限校验保护。">
          <div class="toolbar">
            <span class="toolbar-note">拖动调整顺序；行布局示例：<code>1,2,1,1</code></span>
            <div class="layout-control">
              <span>行布局</span>
              <el-input v-model="layoutPattern" class="layout-input" placeholder="1,2,1" @change="applyLayoutPattern" />
              <el-button size="small" @click="applyLayoutPattern">应用</el-button>
            </div>
            <span class="toolbar-count">已启用 {{ enabledCount }} 项</span>
            <el-button :icon="Plus" :disabled="items.length >= 20" @click="addItem">添加按钮</el-button>
          </div>

          <el-alert v-if="error" class="alert" type="error" :closable="false" show-icon :title="error" />
          <div v-if="items.length === 0" class="empty-state">暂无按钮，点击“添加按钮”开始配置。</div>
          <div
            v-for="(item, index) in items"
            :key="item.id || `new-${index}`"
            class="menu-editor-row"
            :class="{ dragging: dragIndex === index }"
            draggable="true"
            @dragstart="startDrag(index, $event)"
            @dragover.prevent
            @drop="dropItem(index)"
            @dragend="dragIndex = null"
          >
            <div class="row-order" title="拖动调整顺序">
              <el-icon class="drag-handle"><Rank /></el-icon>
              <span class="order-label">{{ index + 1 }}</span>
              <el-button text :icon="ArrowUp" :disabled="index === 0" aria-label="上移" @click="moveItem(index, -1)" />
              <el-button text :icon="ArrowDown" :disabled="index === items.length - 1" aria-label="下移" @click="moveItem(index, 1)" />
            </div>
            <div class="row-fields" :class="{ 'has-custom-action': item.action_type === 'custom' }">
              <el-input v-model="item.label" class="label-input" placeholder="按钮文字" maxlength="64" />
              <el-input v-model="item.icon" class="icon-input" placeholder="Emoji" maxlength="8" />
              <el-select v-model="item.action_type" class="action-type" @change="onActionTypeChange(item)">
                <el-option label="内置动作" value="builtin" />
                <el-option label="HTTPS 链接" value="link" />
                <el-option label="自定义文案" value="custom" />
              </el-select>
              <el-select v-if="item.action_type === 'builtin'" v-model="item.action_key" class="action-key" filterable>
                <el-option v-for="action in builtinActions" :key="action.value" :label="action.label" :value="action.value" />
              </el-select>
              <el-input v-else-if="item.action_type === 'link'" v-model="item.action_value" class="action-value" placeholder="https://..." />
              <div v-else class="custom-fields">
                <label class="field-block">
                  <span class="field-label">点击后发送的文案</span>
                  <el-input v-model="item.message_text" class="action-value" type="textarea" :rows="3" maxlength="4000" show-word-limit placeholder="支持 {name}、{group}、{points} 等变量" />
                </label>
                <div class="link-mode-field">
                  <span class="field-label">链接显示方式</span>
                  <el-radio-group :model-value="item.link_mode" size="small" @change="onLinkModeChange(item, String($event))">
                    <el-radio-button value="text">正文链接</el-radio-button>
                    <el-radio-button value="buttons">内联按钮</el-radio-button>
                  </el-radio-group>
                  <span class="field-hint">内联按钮会显示在消息下方，可添加多个链接。</span>
                </div>
                <label v-if="item.link_mode !== 'buttons'" class="field-block">
                  <span class="field-label">末尾链接（可选）</span>
                  <el-input v-model="item.action_value" class="custom-link" placeholder="https://..." />
                </label>
                <div v-else class="inline-links-editor">
                  <div class="inline-links-toolbar">
                    <label class="inline-columns">
                      <span class="field-label">每行按钮数</span>
                      <el-select :model-value="inlineColumnCount(item)" @change="setInlineColumns(item, Number($event))">
                        <el-option v-for="count in 4" :key="count" :label="`${count} 个`" :value="count" />
                      </el-select>
                    </label>
                    <el-button :icon="Plus" :disabled="item.link_buttons.length >= 10" @click="addInlineLink(item)">添加链接</el-button>
                  </div>
                  <div v-if="item.link_buttons.length === 0" class="inline-links-empty">还没有链接，点击“添加链接”创建消息按钮。</div>
                  <div v-for="(link, linkIndex) in item.link_buttons" :key="linkIndex" class="inline-link-row">
                    <span class="inline-link-index">{{ linkIndex + 1 }}</span>
                    <label class="inline-link-field">
                      <span class="field-label">按钮名称</span>
                      <el-input v-model="link.label" maxlength="64" placeholder="例如：查看商品" />
                    </label>
                    <label class="inline-link-field inline-link-url">
                      <span class="field-label">HTTPS 链接</span>
                      <el-input v-model="link.url" placeholder="https://..." />
                    </label>
                    <div class="inline-link-actions">
                      <el-button text :icon="ArrowUp" :disabled="linkIndex === 0" aria-label="链接上移" @click="moveInlineLink(item, linkIndex, -1)" />
                      <el-button text :icon="ArrowDown" :disabled="linkIndex === item.link_buttons.length - 1" aria-label="链接下移" @click="moveInlineLink(item, linkIndex, 1)" />
                      <el-button text type="danger" :icon="Delete" aria-label="删除链接" @click="removeInlineLink(item, linkIndex)" />
                    </div>
                  </div>
                  <div class="field-hint">最多 10 个按钮。Telegram 的按钮颜色由用户客户端主题决定。</div>
                </div>
              </div>
              <el-switch v-model="item.enabled" active-text="启用" />
              <el-button text type="danger" :icon="Delete" aria-label="删除" @click="removeItem(index)" />
            </div>
          </div>
        </PanelSection>
      </el-col>

      <el-col :xs="24" :lg="8">
        <PanelSection title="Telegram 键盘预览" description="仅启用按钮会显示，实际私聊中按行列排列。">
          <div class="telegram-preview">
            <div class="preview-header">{{ role === "admin" ? "管理员菜单" : "普通用户菜单" }}</div>
            <div v-for="(row, rowIndex) in previewRows" :key="rowIndex" class="preview-row">
              <span v-for="button in row.buttons" :key="button.key" class="preview-button" :style="{ gridColumn: `span ${Math.max(1, Math.floor(4 / row.columns))}` }">{{ button.label }}</span>
            </div>
            <div v-if="previewRows.length === 0" class="preview-empty">暂无启用按钮</div>
          </div>
          <div v-if="customPreview" class="message-preview">
            <div class="preview-header">自定义消息预览</div>
            <div class="preview-message">{{ customPreview.text }}</div>
            <div v-if="customPreview.buttonRows.length" class="inline-button-preview">
              <div v-for="(row, rowIndex) in customPreview.buttonRows" :key="rowIndex" class="inline-button-preview-row">
                <span v-for="(button, buttonIndex) in row" :key="`${rowIndex}-${buttonIndex}`" class="inline-button-preview-item">{{ button.label || "未命名按钮" }}</span>
              </div>
            </div>
          </div>
          <el-divider />
          <div class="preview-rules">
            <div><span>布局</span><strong>每行最多 4 个</strong></div>
            <div><span>显示</span><strong>私聊持久键盘</strong></div>
            <div><span>链接</span><strong>正文链接或多个内联按钮</strong></div>
          </div>
        </PanelSection>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { ArrowDown, ArrowUp, Check, Delete, Plus, Rank, Refresh, RefreshLeft } from "@element-plus/icons-vue";
import PageHeader from "@/components/PageHeader.vue";
import PanelSection from "@/components/PanelSection.vue";
import {
  fetchBotMenu,
  resetBotMenu,
  updateBotMenu,
  type BotMenuInlineLink,
  type BotMenuItem,
  type BotMenuRole,
} from "@/api/botMenu";
import { ApiError } from "@/api/http";

const builtinActions = [
	{ value: "invite_rewards", label: "邀请免费领额度" },
	{ value: "points", label: "我的积分" },
	{ value: "sign", label: "每日签到" },
	{ value: "rank", label: "排行榜" },
	{ value: "exchange", label: "积分兑换免费额度" },
	{ value: "purchase", label: "直接购买额度" },
	{ value: "shop", label: "小铺地址" },
  { value: "lottery", label: "抽奖大厅" },
  { value: "daily_lottery", label: "每日额度抽奖" },
  { value: "help", label: "使用帮助" },
  { value: "info", label: "会话信息" },
  { value: "private_console", label: "运营工作台" },
  { value: "admin_center", label: "群管中心" },
  { value: "admin_config", label: "群组配置" },
  { value: "scheduled_posts", label: "定时发帖" },
  { value: "hide_keyboard", label: "收起私聊菜单" },
];

type EditableBotMenuItem = BotMenuItem & { inline_columns: number };

const role = ref<BotMenuRole>("member");
const items = ref<EditableBotMenuItem[]>([]);
const loading = ref(false);
const saving = ref(false);
const error = ref("");
const dragIndex = ref<number | null>(null);
const layoutPattern = ref("2");

const enabledCount = computed(() => items.value.filter((item) => item.enabled).length);
const customPreview = computed(() => {
  const item = items.value.find((candidate) => candidate.enabled && candidate.action_type === "custom" && candidate.message_text.trim());
  if (!item) return null;
  const text = item.message_text.trim()
    .replaceAll("{name}", "示例用户")
    .replaceAll("{group}", "示例群组")
    .replaceAll("{points}", "20")
    .replaceAll("{today_invite}", "3")
    .replaceAll("{successful_invites}", "5")
    .replaceAll("{exchanged_amount}", "10");
  if (item.link_mode !== "buttons") {
    return {
      text: item.action_value.trim() ? `${text}\n\n链接：${item.action_value.trim()}` : text,
      buttonRows: [] as BotMenuInlineLink[][],
    };
  }
  return { text, buttonRows: groupInlineLinks(item.link_buttons) };
});
const previewRows = computed(() => {
  const rows = new Map<number, EditableBotMenuItem[]>();
  items.value.filter((item) => item.enabled).forEach((item) => {
    const row = rows.get(item.row_index) || [];
    row.push(item);
    rows.set(item.row_index, row);
  });
  return [...rows.entries()].sort((a, b) => a[0] - b[0]).map(([, row]) => {
    const buttons = row.sort((a, b) => a.column_index - b.column_index).map((item, index) => ({
      key: item.id || `${item.action_key}-${index}`,
      label: renderLabel(item),
    }));
    return { buttons, columns: Math.max(1, buttons.length) };
  });
});

function renderLabel(item: BotMenuItem): string {
  const label = item.label.trim();
  const icon = item.icon.trim();
  return icon && !label.startsWith(icon) ? `${icon} ${label}` : label;
}

function normalizePositions(): void {
  const pattern = parseLayoutPattern();
  let row = 0;
  let column = 0;
  items.value.forEach((item, index) => {
    if (index === 0) {
      row = 0;
      column = 0;
    }
    item.row_index = row;
    item.column_index = column;
    column += 1;
    const capacity = pattern[Math.min(row, pattern.length - 1)];
    if (column >= capacity) {
      row += 1;
      column = 0;
    }
  });
}

function parseLayoutPattern(): number[] {
  const values = layoutPattern.value.split(",").map((value) => Number(value.trim()));
  const valid = values.filter((value) => Number.isInteger(value) && value >= 1 && value <= 4);
  return valid.length ? valid : [2];
}

function applyLayoutPattern(): void {
  const values = layoutPattern.value.split(",").map((value) => Number(value.trim()));
  if (!values.length || values.some((value) => !Number.isInteger(value) || value < 1 || value > 4)) {
    ElMessage.warning("行布局只能填写 1–4 的数字，例如 1,2,1");
    layoutPattern.value = "2";
    return;
  }
  normalizePositions();
}

function cloneItems(next: BotMenuItem[]): void {
  items.value = next
    .slice()
    .sort((a, b) => a.row_index - b.row_index || a.column_index - b.column_index)
    .map((item) => ({
      ...item,
      link_mode: item.link_mode === "buttons" ? "buttons" : "text",
      link_buttons: Array.isArray(item.link_buttons)
        ? item.link_buttons
          .slice()
          .sort((a, b) => a.row_index - b.row_index || a.column_index - b.column_index)
          .map((link) => ({ ...link }))
        : [],
      inline_columns: deriveInlineColumns(item.link_buttons),
    }));
  const rowCounts = new Map<number, number>();
  items.value.forEach((item) => rowCounts.set(item.row_index, (rowCounts.get(item.row_index) || 0) + 1));
  const counts = [...rowCounts.values()];
  layoutPattern.value = counts.length ? counts.join(",") : "2";
}

async function loadMenu(): Promise<void> {
  loading.value = true;
  error.value = "";
  try {
    const result = await fetchBotMenu(role.value);
    cloneItems(result.items);
  } catch {
    items.value = [];
    error.value = "菜单加载失败，请检查后台权限和机器人配置。";
  } finally {
    loading.value = false;
  }
}

function addItem(): void {
  if (items.value.length >= 20) return;
  items.value.push({
    role: role.value,
    label: "新按钮",
    icon: "",
    action_type: "builtin",
    action_key: "help",
    action_value: "",
    message_text: "",
    link_mode: "text",
    link_buttons: [],
    inline_columns: 1,
    row_index: 0,
    column_index: 0,
    enabled: true,
  });
  normalizePositions();
}

function removeItem(index: number): void {
  items.value.splice(index, 1);
}

function moveItem(index: number, offset: number): void {
  const next = index + offset;
  if (next < 0 || next >= items.value.length) return;
  const [item] = items.value.splice(index, 1);
  items.value.splice(next, 0, item);
  normalizePositions();
}

function startDrag(index: number, event: DragEvent): void {
  dragIndex.value = index;
  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = "move";
    event.dataTransfer.setData("text/plain", String(index));
  }
}

function dropItem(targetIndex: number): void {
  if (dragIndex.value == null || dragIndex.value === targetIndex) return;
  const [item] = items.value.splice(dragIndex.value, 1);
  items.value.splice(targetIndex, 0, item);
  dragIndex.value = null;
  normalizePositions();
}

function onActionTypeChange(item: EditableBotMenuItem): void {
  if (item.action_type === "builtin") {
    item.action_value = "";
    item.message_text = "";
    item.link_mode = "text";
    item.link_buttons = [];
    item.inline_columns = 1;
  } else if (item.action_type === "link") {
    item.action_key = "";
    item.message_text = "";
    item.link_mode = "text";
    item.link_buttons = [];
    item.inline_columns = 1;
  } else {
    item.action_key = "";
    item.link_mode = "text";
    item.link_buttons = [];
    item.inline_columns = 1;
  }
}

function onLinkModeChange(item: EditableBotMenuItem, value: string): void {
  item.link_mode = value === "buttons" ? "buttons" : "text";
  if (item.link_mode === "buttons" && item.link_buttons.length === 0) addInlineLink(item);
}

function deriveInlineColumns(links: BotMenuInlineLink[] | undefined): number {
  if (!Array.isArray(links) || !links.length) return 1;
  const rowCounts = new Map<number, number>();
  links.forEach((link) => rowCounts.set(link.row_index, (rowCounts.get(link.row_index) || 0) + 1));
  return Math.max(1, Math.min(4, ...rowCounts.values()));
}

function inlineColumnCount(item: EditableBotMenuItem): number {
  return Math.max(1, Math.min(4, item.inline_columns || 1));
}

function normalizeInlineLinks(item: EditableBotMenuItem, columns = inlineColumnCount(item)): void {
  const safeColumns = Math.max(1, Math.min(4, columns));
  item.inline_columns = safeColumns;
  item.link_buttons.forEach((link, index) => {
    link.row_index = Math.floor(index / safeColumns);
    link.column_index = index % safeColumns;
  });
}

function setInlineColumns(item: EditableBotMenuItem, columns: number): void {
  normalizeInlineLinks(item, columns);
}

function addInlineLink(item: EditableBotMenuItem): void {
  if (item.link_buttons.length >= 10) return;
  const columns = inlineColumnCount(item);
  item.link_buttons.push({ label: "查看链接", url: "", row_index: 0, column_index: 0 });
  normalizeInlineLinks(item, columns);
}

function removeInlineLink(item: EditableBotMenuItem, index: number): void {
  const columns = inlineColumnCount(item);
  item.link_buttons.splice(index, 1);
  normalizeInlineLinks(item, columns);
}

function moveInlineLink(item: EditableBotMenuItem, index: number, offset: number): void {
  const next = index + offset;
  if (next < 0 || next >= item.link_buttons.length) return;
  const columns = inlineColumnCount(item);
  const [link] = item.link_buttons.splice(index, 1);
  item.link_buttons.splice(next, 0, link);
  normalizeInlineLinks(item, columns);
}

function groupInlineLinks(links: BotMenuInlineLink[]): BotMenuInlineLink[][] {
  const rows = new Map<number, BotMenuInlineLink[]>();
  links
    .slice()
    .sort((a, b) => a.row_index - b.row_index || a.column_index - b.column_index)
    .forEach((link) => {
      const row = rows.get(link.row_index) || [];
      row.push(link);
      rows.set(link.row_index, row);
    });
  return [...rows.entries()].sort((a, b) => a[0] - b[0]).map(([, row]) => row);
}

async function saveMenu(): Promise<void> {
  if (items.value.some((item) => !item.label.trim())) {
    ElMessage.warning("请填写所有按钮文字");
    return;
  }
  if (items.value.some((item) => item.action_type === "link" && !/^https:\/\//i.test(item.action_value.trim()))) {
    ElMessage.warning("链接按钮只支持 https:// 地址");
    return;
  }
  if (items.value.some((item) => item.action_type === "custom" && !item.message_text.trim())) {
    ElMessage.warning("自定义动作必须填写发送文案");
    return;
  }
  if (items.value.some((item) => item.action_type === "custom" && item.link_mode !== "buttons" && item.action_value.trim() && !/^https:\/\//i.test(item.action_value.trim()))) {
    ElMessage.warning("自定义动作链接只支持 https:// 地址");
    return;
  }
  const buttonItems = items.value.filter((item) => item.action_type === "custom" && item.link_mode === "buttons");
  if (buttonItems.some((item) => item.link_buttons.length === 0 || item.link_buttons.length > 10)) {
    ElMessage.warning("内联按钮模式需要添加 1–10 个链接");
    return;
  }
  if (buttonItems.some((item) => item.link_buttons.some((link) => !link.label.trim()))) {
    ElMessage.warning("请填写所有内联按钮名称");
    return;
  }
  if (buttonItems.some((item) => item.link_buttons.some((link) => !/^https:\/\//i.test(link.url.trim())))) {
    ElMessage.warning("内联按钮只支持 https:// 地址");
    return;
  }
  const labels = new Set<string>();
  for (const item of items.value) {
    const label = renderLabel(item);
    if (labels.has(label)) {
      ElMessage.warning(`按钮文字重复：${label}`);
      return;
    }
    labels.add(label);
  }
  if (enabledCount.value > 20) {
    ElMessage.warning("启用按钮不能超过 20 个");
    return;
  }
  normalizePositions();
  buttonItems.forEach((item) => normalizeInlineLinks(item));
  saving.value = true;
  try {
    const payload = items.value.map(({ inline_columns: _inlineColumns, ...item }) => item);
    const result = await updateBotMenu(role.value, payload);
    cloneItems(result.items);
    ElMessage.success("菜单已保存，刷新私聊菜单后生效");
  } catch (err) {
    ElMessage.error(apiErrorMessage(err));
  } finally {
    saving.value = false;
  }
}

function apiErrorMessage(err: unknown): string {
  if (err instanceof ApiError) {
    const payload = err.payload as { error?: string } | undefined;
    if (payload?.error) return `保存失败：${payload.error}`;
  }
  return "菜单保存失败，请检查按钮名称和动作参数";
}

async function restoreDefaults(): Promise<void> {
  try {
    await ElMessageBox.confirm("将覆盖当前角色的菜单配置，是否继续？", "恢复默认", { type: "warning" });
    saving.value = true;
    const result = await resetBotMenu(role.value);
    cloneItems(result.items);
    ElMessage.success("已恢复默认菜单");
  } catch (err) {
    if (err !== "cancel" && err !== "close") ElMessage.error("恢复默认失败");
  } finally {
    saving.value = false;
  }
}

onMounted(loadMenu);
</script>

<style scoped>
.toolbar { display: flex; align-items: center; flex-wrap: wrap; justify-content: space-between; gap: 10px; margin-bottom: 14px; }
.toolbar-note { color: var(--app-muted); font-size: 12px; }
.toolbar-count { color: var(--app-muted); font-size: 12px; white-space: nowrap; }
.layout-control { display: inline-flex; align-items: center; gap: 6px; color: var(--app-muted); font-size: 12px; }
.layout-input { width: 92px; }
.menu-editor-row { display: flex; gap: 8px; align-items: center; width: 100%; max-width: 100%; min-width: 0; padding: 12px 0; border-bottom: 1px solid var(--app-border); transition: opacity .15s ease, background .15s ease; }
.menu-editor-row.dragging { opacity: .45; background: var(--app-tint-light); }
.menu-editor-row:last-child { border-bottom: 0; }
.row-order { display: flex; align-items: center; gap: 2px; color: var(--app-muted); }
.drag-handle { cursor: grab; color: var(--app-accent); }
.drag-handle:active { cursor: grabbing; }
.order-label { width: 22px; text-align: center; font-size: 12px; }
.row-fields { display: grid; grid-template-columns: minmax(0, 1.2fr) 68px 108px minmax(0, 1fr) auto 28px; gap: 7px; align-items: center; width: 0; flex: 1 1 auto; min-width: 0; }
.label-input, .action-key, .action-value { min-width: 0; }
.row-fields.has-custom-action .custom-fields { grid-column: 1 / -1; grid-row: 2; }
.row-fields.has-custom-action > :nth-child(5) { grid-column: 5; grid-row: 1; }
.row-fields.has-custom-action > :nth-child(6) { grid-column: 6; grid-row: 1; }
.custom-fields { display: grid; gap: 12px; min-width: 0; padding: 12px; border: 1px solid var(--app-border); border-radius: 8px; background: var(--app-surface-2); }
.field-block, .inline-link-field, .inline-columns { display: grid; gap: 5px; min-width: 0; }
.field-label { color: var(--app-text); font-size: 12px; font-weight: 600; line-height: 1.4; }
.field-hint { color: var(--app-muted); font-size: 12px; line-height: 1.5; }
.link-mode-field { display: flex; flex-wrap: wrap; align-items: center; gap: 8px 12px; }
.inline-links-editor { display: grid; gap: 10px; min-width: 0; }
.inline-links-toolbar { display: flex; align-items: end; justify-content: space-between; gap: 12px; }
.inline-columns { width: 132px; }
.inline-link-row { display: grid; grid-template-columns: 24px minmax(120px, .7fr) minmax(220px, 1.3fr) auto; gap: 8px; align-items: end; min-width: 0; }
.inline-link-index { align-self: center; color: var(--app-muted); font-size: 12px; text-align: center; }
.inline-link-actions { display: flex; align-items: center; min-height: 32px; }
.inline-links-empty { padding: 14px 12px; border: 1px dashed var(--app-border); border-radius: 6px; color: var(--app-muted); text-align: center; font-size: 12px; }
.icon-input { width: 72px; }
.action-type { width: 116px; }
.row-fields :deep(.el-input), .row-fields :deep(.el-select), .row-fields :deep(.el-input-number) { max-width: 100%; min-width: 0; }
.empty-state, .preview-empty { color: var(--app-muted); text-align: center; padding: 28px 12px; font-size: 13px; }
.telegram-preview { padding: 12px; border: 1px solid var(--app-border); border-radius: 10px; background: var(--app-surface-2); }
.preview-header { margin-bottom: 12px; color: var(--app-muted); font-size: 12px; }
.preview-row { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 6px; margin-bottom: 6px; }
.preview-button { min-width: 0; overflow: hidden; padding: 9px 6px; border: 1px solid var(--app-border); border-radius: 7px; background: var(--app-surface); text-align: center; text-overflow: ellipsis; white-space: nowrap; font-size: 12px; }
.message-preview { margin-top: 14px; padding: 12px; border: 1px solid var(--app-border); border-radius: 8px; background: var(--app-surface-2); }
.preview-message { overflow-wrap: anywhere; color: var(--app-text); font-size: 13px; line-height: 1.6; white-space: pre-wrap; }
.inline-button-preview { display: grid; gap: 6px; margin-top: 10px; }
.inline-button-preview-row { display: flex; gap: 6px; }
.inline-button-preview-item { flex: 1 1 0; min-width: 0; overflow: hidden; padding: 8px 6px; border-radius: 6px; background: var(--app-accent); color: #fff; text-align: center; text-overflow: ellipsis; white-space: nowrap; font-size: 12px; }
.preview-rules { display: grid; gap: 9px; color: var(--app-muted); font-size: 12px; }
.preview-rules div { display: flex; justify-content: space-between; gap: 8px; }
.preview-rules strong { color: var(--app-text); font-weight: 600; }
@media (max-width: 1100px) { .menu-editor-row { align-items: flex-start; } .row-fields { grid-template-columns: minmax(100px, 1fr) 68px 108px minmax(110px, 1fr) auto 28px; } }
@media (max-width: 760px) {
  .menu-editor-row { align-items: flex-start; }
  .row-order { flex-direction: column; }
  .row-fields { grid-template-columns: minmax(0, 1fr) 68px; }
  .action-type, .action-key, .action-value { width: 100%; }
  .row-fields > :nth-child(3), .row-fields > :nth-child(4), .row-fields > :nth-child(5), .row-fields > :nth-child(6) { grid-column: span 2; }
  .row-fields.has-custom-action .custom-fields { grid-column: 1 / -1; grid-row: auto; }
  .row-fields.has-custom-action > :nth-child(5), .row-fields.has-custom-action > :nth-child(6) { grid-column: span 2; grid-row: auto; }
  .inline-links-toolbar { align-items: stretch; flex-direction: column; }
  .inline-columns { width: 100%; }
  .inline-link-row { grid-template-columns: 24px minmax(0, 1fr); align-items: start; }
  .inline-link-url, .inline-link-actions { grid-column: 2; }
  .inline-link-index { grid-row: 1 / span 3; }
  .inline-link-actions { justify-content: flex-end; }
}
:deep(.panel-body) { overflow-x: hidden; }
</style>
