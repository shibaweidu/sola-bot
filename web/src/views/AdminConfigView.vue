<template>
  <div class="page">
    <PageHeader eyebrow="群组管理" title="群组配置" description="统一管理新人验证、欢迎消息和垃圾消息拦截。">
      <template #actions>
        <ChatSelect v-model="selectedChatId" @update:model-value="loadConfig" />
        <el-button :icon="Refresh" :loading="loading" @click="loadConfig">刷新</el-button>
        <el-button type="primary" :icon="Check" :loading="saving" @click="submitConfig">保存</el-button>
      </template>
    </PageHeader>

    <el-row :gutter="16">
      <el-col :xs="24" :lg="16">
        <PanelSection title="新人欢迎" description="验证关闭或验证通过后发送，0 秒表示不自动删除。">
          <el-form label-position="top">
            <el-form-item label="启用欢迎消息">
              <el-switch v-model="form.welcome_enabled" />
            </el-form-item>
            <el-form-item label="欢迎语">
              <el-input v-model="form.welcome_text" type="textarea" :rows="3" placeholder="欢迎 {name} 加入！" />
            </el-form-item>
            <el-form-item label="自动删除（秒）">
              <el-input-number v-model="form.welcome_delete_seconds" class="wide-control" :min="0" :max="3600" />
            </el-form-item>
          </el-form>
        </PanelSection>

        <PanelSection title="强制订阅" description="未订阅指定频道的成员将被禁言或移出群组；管理员和验证白名单自动豁免。">
          <el-form label-position="top">
            <el-form-item label="启用强制订阅">
              <el-switch v-model="form.force_subscribe_enabled" />
            </el-form-item>
            <el-form-item label="必订阅频道">
              <el-input
                v-model="form.force_subscribe_channels"
                type="textarea"
                :rows="4"
                placeholder="一行一个：@example_channel、https://t.me/example_channel 或 -1001234567890"
              />
            </el-form-item>
            <el-form-item label="频道显示名称（可选）">
              <el-input
                v-model="form.force_subscribe_channel_labels"
                type="textarea"
                :rows="3"
                placeholder="一行一个：@example_channel | 示例频道"
              />
              <div class="field-hint">只改变按钮和提示中的显示文字，不改变 Telegram 校验目标。</div>
            </el-form-item>
            <el-form-item label="未订阅提示文案">
              <el-input
                v-model="form.force_subscribe_message"
                type="textarea"
                :rows="3"
                placeholder="{name}，请先订阅指定频道后才能发言。\n待订阅：{channels}"
              />
              <div class="field-hint">支持变量：`{name}` 用户名，`{channels}` 待订阅频道显示名。</div>
            </el-form-item>
            <el-form-item label="移出群组提示文案">
              <el-input
                v-model="form.force_subscribe_kick_message"
                type="textarea"
                :rows="2"
                placeholder="{name} 未完成必需频道订阅，已移出群组。\n待订阅：{channels}"
              />
            </el-form-item>
            <el-form-item label="未订阅处理">
              <el-radio-group v-model="form.force_subscribe_action">
                <el-radio-button label="mute">禁言，订阅后解除</el-radio-button>
                <el-radio-button label="kick">移出群组</el-radio-button>
              </el-radio-group>
            </el-form-item>
          </el-form>
        </PanelSection>

        <PanelSection title="新人验证" description="支持按钮、数字验证码、选择题、投票、数学题和 Turnstile。">
          <el-form label-position="top">
            <el-row :gutter="12">
              <el-col :xs="24" :md="6">
                <el-form-item label="入群验证">
                  <el-switch v-model="form.verify_enabled" />
                </el-form-item>
              </el-col>
              <el-col :xs="24" :md="6">
                <el-form-item label="验证方式">
                  <el-select v-model="form.verify_type" class="wide-control">
                    <el-option label="按钮确认" value="button" />
                    <el-option label="数字验证码" value="captcha" />
                    <el-option label="选择题" value="multi_choice" />
                    <el-option label="Telegram 投票" value="poll" />
                    <el-option label="数学题" value="math" />
                    <el-option label="Turnstile" value="turnstile" />
                  </el-select>
                </el-form-item>
              </el-col>
              <el-col :xs="24" :md="6">
                <el-form-item label="验证超时">
                  <el-input-number v-model="form.verify_timeout" class="wide-control" :min="10" :max="3600" />
                </el-form-item>
              </el-col>
              <el-col :xs="24" :md="6">
                <el-form-item label="难度">
                  <el-select v-model="form.verify_difficulty" class="wide-control">
                    <el-option label="宽松" value="easy" />
                    <el-option label="标准" value="medium" />
                    <el-option label="严格" value="hard" />
                  </el-select>
                </el-form-item>
              </el-col>
            </el-row>
            <template v-if="form.verify_type === 'multi_choice' || form.verify_type === 'poll'">
              <el-form-item label="验证问题">
                <el-input v-model="form.verify_question" />
              </el-form-item>
              <el-row :gutter="12">
                <el-col :xs="24" :md="18">
                  <el-form-item label="选项（JSON 数组）">
                    <el-input v-model="form.verify_options" placeholder='["选项一","选项二"]' />
                  </el-form-item>
                </el-col>
                <el-col :xs="24" :md="6">
                  <el-form-item label="正确项序号">
                    <el-input-number v-model="form.verify_correct_index" class="wide-control" :min="0" :max="9" />
                  </el-form-item>
                </el-col>
              </el-row>
            </template>
          </el-form>
        </PanelSection>

        <PanelSection title="垃圾消息拦截" description="管理员自动豁免；达到分值后按风险等级删除、警告、禁言或封禁。">
          <el-form label-position="top">
            <el-row :gutter="12">
              <el-col :xs="12" :md="6">
                <el-form-item label="关键词规则">
                  <el-switch v-model="form.keyword_filter_enabled" />
                </el-form-item>
              </el-col>
              <el-col :xs="12" :md="6">
                <el-form-item label="拦截链接">
                  <el-switch v-model="form.block_links" />
                </el-form-item>
              </el-col>
              <el-col :xs="12" :md="6">
                <el-form-item label="拦截转发">
                  <el-switch v-model="form.block_forwards" />
                </el-form-item>
              </el-col>
              <el-col :xs="12" :md="6">
                <el-form-item label="拦截媒体">
                  <el-switch v-model="form.block_media" />
                </el-form-item>
              </el-col>
            </el-row>
            <el-row :gutter="12">
              <el-col :xs="24" :md="8">
                <el-form-item label="垃圾分数阈值">
                  <el-input-number v-model="form.spam_score_threshold" class="wide-control" :min="1" :max="100" />
                </el-form-item>
              </el-col>
              <el-col :xs="12" :md="8">
                <el-form-item label="AI 二次判定">
                  <el-switch v-model="form.ai_filter_enabled" />
                </el-form-item>
              </el-col>
              <el-col :xs="12" :md="8">
                <el-form-item label="限制未验证成员">
                  <el-switch v-model="form.restrict_unverified" />
                </el-form-item>
              </el-col>
            </el-row>
            <el-row :gutter="12" v-if="form.block_links">
              <el-col :xs="24" :md="12">
                <el-form-item label="链接白名单（一行一个域名）">
                  <el-input v-model="form.link_whitelist" type="textarea" :rows="3" placeholder="example.com" />
                </el-form-item>
              </el-col>
              <el-col :xs="24" :md="12">
                <el-form-item label="链接黑名单（一行一个域名）">
                  <el-input v-model="form.link_blacklist" type="textarea" :rows="3" placeholder="spam.example" />
                </el-form-item>
              </el-col>
            </el-row>
            <el-form-item label="累计警告封禁上限">
              <el-input-number v-model="form.warn_limit" class="wide-control" :min="1" :max="20" />
            </el-form-item>
          </el-form>
        </PanelSection>
      </el-col>

      <el-col :xs="24" :lg="8">
        <PanelSection title="预览" description="保存前检查最终下发效果。">
          <div class="preview">
            <div><span>群组 ID</span><strong>{{ selectedChatId || "-" }}</strong></div>
            <div><span>欢迎消息</span><strong>{{ form.welcome_enabled ? "开启" : "关闭" }}</strong></div>
            <div><span>验证</span><strong>{{ form.verify_enabled ? "开启" : "关闭" }}</strong></div>
            <div><span>验证方式</span><strong>{{ verifyTypeLabel }}</strong></div>
            <div><span>强制订阅</span><strong>{{ form.force_subscribe_enabled ? form.force_subscribe_action === "mute" ? "禁言" : "移出" : "关闭" }}</strong></div>
            <div><span>超时</span><strong>{{ form.verify_timeout }} 秒</strong></div>
            <div><span>垃圾阈值</span><strong>{{ form.spam_score_threshold }}</strong></div>
            <div><span>警告上限</span><strong>{{ form.warn_limit }}</strong></div>
            <div class="preview-message">{{ form.welcome_text }}</div>
          </div>
        </PanelSection>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { ElMessage } from "element-plus";
import { Check, Refresh } from "@element-plus/icons-vue";
import { useRoute } from "vue-router";
import ChatSelect from "@/components/ChatSelect.vue";
import PageHeader from "@/components/PageHeader.vue";
import PanelSection from "@/components/PanelSection.vue";
import { fetchAdminConfig, updateAdminConfig } from "@/api/admin";
import type { ChatAdminConfigPayload } from "@/types/api";

const route = useRoute();
const selectedChatId = ref("");
const loading = ref(false);
const saving = ref(false);
const form = reactive<ChatAdminConfigPayload>({
  welcome_text: "",
  welcome_enabled: true,
  welcome_delete_seconds: 30,
  verify_enabled: false,
  verify_type: "button",
  verify_timeout: 60,
  verify_question: "",
  verify_options: "[]",
  verify_correct_index: 0,
  verify_difficulty: "medium",
  warn_limit: 3,
  force_subscribe_enabled: false,
  force_subscribe_channels: "",
  force_subscribe_action: "mute",
  force_subscribe_message: "",
  force_subscribe_kick_message: "",
  force_subscribe_channel_labels: "",
  block_links: false,
  link_whitelist: "",
  link_blacklist: "",
  block_forwards: false,
  block_media: false,
  keyword_filter_enabled: true,
  spam_score_threshold: 60,
  ai_filter_enabled: false,
  restrict_unverified: true,
});

const verifyTypeLabel = computed(() => ({
  button: "按钮确认",
  captcha: "数字验证码",
  multi_choice: "选择题",
  poll: "Telegram 投票",
  math: "数学题",
  turnstile: "Turnstile",
}[form.verify_type]));

async function loadConfig(): Promise<void> {
  if (!selectedChatId.value) return;
  loading.value = true;
  try {
    Object.assign(form, await fetchAdminConfig(selectedChatId.value));
  } catch {
    ElMessage.error("服务暂时不可用");
  } finally {
    loading.value = false;
  }
}

async function submitConfig(): Promise<void> {
  if (!selectedChatId.value) return;
  saving.value = true;
  try {
    Object.assign(form, await updateAdminConfig(selectedChatId.value, { ...form }));
    ElMessage.success("群组配置已保存");
  } catch {
    ElMessage.error("服务暂时不可用");
  } finally {
    saving.value = false;
  }
}

onMounted(() => {
  const queryChatID = route.query.chat_id;
  if (typeof queryChatID === "string") {
    selectedChatId.value = queryChatID;
  }
  void loadConfig();
});
</script>

<style scoped>
.page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.wide-control {
  width: 100%;
}

.field-hint {
  margin-top: 6px;
  color: var(--app-muted);
  font-size: 12px;
  line-height: 1.5;
}

.preview {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.preview div {
  display: flex;
  justify-content: space-between;
  gap: 12px;
}

.preview span {
  color: var(--app-muted);
}

.preview-message {
  padding: 12px;
  border: 1px solid var(--app-border);
  border-radius: 8px;
  line-height: 1.6;
  background: var(--app-table-header-bg);
}

@media (max-width: 720px) {
}
</style>
