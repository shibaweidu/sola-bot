import { request } from "@/api/http";
import type { ChatID } from "@/types/api";

export interface PointCenterConfig {
  chat_id: ChatID;
  invite_enabled: boolean;
  inviter_reward: number;
  invitee_reward: number;
  sign_enabled: boolean;
  sign_reward: number;
  exchange_enabled: boolean;
  exchange_minimum: number;
  exchange_rate: number;
  exchange_url: string;
  exchange_instructions: string;
  purchase_url: string;
  purchase_text: string;
  shop_url: string;
  shop_text: string;
  invite_text: string;
  invite_page_template: string;
  invite_join_url: string;
  invite_join_text: string;
  invite_success_text: string;
  points_text: string;
  sign_text: string;
  exchange_text: string;
  rank_text: string;
}

export interface ExchangeCodeRecord {
  id: string;
  code: string;
  batch_name: string;
  amount: number;
  redeem_url: string;
  expires_at?: string;
  status: string;
  assigned_user: number;
  assigned_chat: number;
  assigned_at?: string;
}

export interface ExchangeCodeSummary {
  amount: number;
  available: number;
  assigned: number;
  used: number;
  total: number;
}

export function fetchPointCenterConfig(chatId: ChatID): Promise<PointCenterConfig> {
  return request<PointCenterConfig>(`/point-center/config?chat_id=${encodeURIComponent(String(chatId))}`);
}

export function updatePointCenterConfig(payload: PointCenterConfig): Promise<PointCenterConfig> {
  return request<PointCenterConfig>("/point-center/config", { method: "PUT", body: payload });
}

export function resetReferralForTesting(chatId: ChatID, userId: number): Promise<{ ok: boolean }> {
	return request<{ ok: boolean }>("/point-center/referral/reset", {
		method: "POST",
		body: { chat_id: Number(chatId), user_id: userId },
	});
}

export interface InventoryPage<T> { items: T[]; total: number; page: number; page_size: number }
export interface InventoryBatchResult { updated: number; deleted: number; skipped: number }

export function fetchExchangeCodes(status = "", amount?: number, page = 1, pageSize = 20): Promise<InventoryPage<ExchangeCodeRecord>> {
  const params = new URLSearchParams();
  if (status) params.set("status", status);
  if (amount !== undefined) params.set("amount", String(amount));
  params.set("page", String(page));
  params.set("page_size", String(pageSize));
  const suffix = params.toString() ? `?${params.toString()}` : "";
  return request<InventoryPage<ExchangeCodeRecord>>(`/point-center/codes${suffix}`);
}

export function batchUpdateExchangeCodes(payload: { ids: string[]; redeem_url?: string; batch_name?: string; expires_at?: string }): Promise<InventoryBatchResult> {
  return request<InventoryBatchResult>("/point-center/codes/batch", { method: "PATCH", body: payload });
}

export function batchDeleteExchangeCodes(ids: string[]): Promise<InventoryBatchResult> {
  return request<InventoryBatchResult>("/point-center/codes/batch", { method: "DELETE", body: { ids } });
}

export function fetchExchangeCodeSummary(): Promise<{ items: ExchangeCodeSummary[] }> {
  return request<{ items: ExchangeCodeSummary[] }>("/point-center/codes/summary");
}

export function importExchangeCodes(payload: {
  batch_name: string;
  amount: number;
  redeem_url: string;
  expires_at?: string;
  codes: string[];
}): Promise<{ imported: number; skipped: number }> {
  return request<{ imported: number; skipped: number }>("/point-center/codes/import", {
    method: "POST",
    body: payload,
  });
}
