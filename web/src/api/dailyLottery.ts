import { request } from "@/api/http";
import type { ChatID } from "@/types/api";

export interface DailyLotteryConfig {
  chat_id: ChatID;
  enabled: boolean;
  daily_attempts: number;
  cost_points: number;
  paid_enabled: boolean;
  guarantee_on_last: boolean;
}

export interface DailyLotteryPrize {
  id?: number;
  chat_id: ChatID;
  amount: number;
  weight: number;
  enabled: boolean;
  available_code: number;
}

export interface DailyLotteryCode {
  id: string;
  chat_id: ChatID;
  code: string;
  amount: number;
  batch_name: string;
  redeem_url: string;
  expires_at?: string;
  status: string;
  assigned_user: number;
  assigned_at?: string;
}

export interface DailyLotteryCodeSummary {
  amount: number;
  available: number;
  assigned: number;
  used: number;
  total: number;
}

export function fetchDailyLotteryConfig(chatId: ChatID): Promise<DailyLotteryConfig> {
  return request<DailyLotteryConfig>(`/point-center/daily-lottery/config?chat_id=${encodeURIComponent(String(chatId))}`);
}

export function fetchDailyLotteryPrizes(chatId: ChatID): Promise<{ items: DailyLotteryPrize[] }> {
  return request<{ items: DailyLotteryPrize[] }>(`/point-center/daily-lottery/prizes?chat_id=${encodeURIComponent(String(chatId))}`);
}

export function updateDailyLottery(payload: {
  chat_id: ChatID;
  enabled: boolean;
  cost_points: number;
  paid_enabled: boolean;
  guarantee_on_last: boolean;
  prizes: Array<{ amount: number; weight: number; enabled: boolean }>;
}): Promise<DailyLotteryConfig> {
  return request<DailyLotteryConfig>("/point-center/daily-lottery/config", { method: "PUT", body: payload });
}

export function fetchDailyLotteryCodes(chatId: ChatID, amount?: number, status?: string): Promise<{ items: DailyLotteryCode[] }> {
  const params = new URLSearchParams({ chat_id: String(chatId) });
  if (amount !== undefined) params.set("amount", String(amount));
  if (status) params.set("status", status);
  return request<{ items: DailyLotteryCode[] }>(`/point-center/daily-lottery/codes?${params.toString()}`);
}

export function fetchDailyLotteryCodeSummary(chatId: ChatID): Promise<{ items: DailyLotteryCodeSummary[] }> {
  return request<{ items: DailyLotteryCodeSummary[] }>(`/point-center/daily-lottery/codes/summary?chat_id=${encodeURIComponent(String(chatId))}`);
}

export function importDailyLotteryCodes(payload: {
  chat_id: ChatID;
  amount: number;
  batch_name: string;
  redeem_url: string;
  expires_at?: string;
  codes: string[];
}): Promise<{ imported: number; skipped: number }> {
  return request<{ imported: number; skipped: number }>("/point-center/daily-lottery/codes/import", { method: "POST", body: payload });
}

export function resetDailyLotteryAttempts(payload: { chat_id: ChatID; user_id: number }): Promise<{ ok: boolean }> {
  return request<{ ok: boolean }>("/point-center/daily-lottery/reset", { method: "POST", body: payload });
}
