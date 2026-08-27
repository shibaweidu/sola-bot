import { request } from "@/api/http";
import type { ChatID } from "@/types/api";

export interface DailyLotteryConfig {
  chat_id: ChatID;
  enabled: boolean;
  daily_attempts: number;
  cost_points: number;
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
  guarantee_on_last: boolean;
  prizes: Array<{ amount: number; weight: number; enabled: boolean }>;
}): Promise<DailyLotteryConfig> {
  return request<DailyLotteryConfig>("/point-center/daily-lottery/config", { method: "PUT", body: payload });
}
