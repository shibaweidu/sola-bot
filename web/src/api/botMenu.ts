import { request } from "@/api/http";

export type BotMenuRole = "member" | "admin";
export type BotMenuActionType = "builtin" | "link" | "custom";
export type BotMenuLinkMode = "text" | "buttons";

export interface BotMenuInlineLink {
  label: string;
  url: string;
  row_index: number;
  column_index: number;
}

export interface BotMenuItem {
  id?: string;
  role: BotMenuRole;
  label: string;
  icon: string;
  action_type: BotMenuActionType;
  action_key: string;
  action_value: string;
  message_text: string;
  link_mode: BotMenuLinkMode;
  link_buttons: BotMenuInlineLink[];
  row_index: number;
  column_index: number;
  enabled: boolean;
}

export interface BotMenuResponse {
  role: BotMenuRole;
  items: BotMenuItem[];
}

export function fetchBotMenu(role: BotMenuRole): Promise<BotMenuResponse> {
  return request<BotMenuResponse>(`/bot-menu?role=${role}`);
}

export function updateBotMenu(role: BotMenuRole, items: BotMenuItem[]): Promise<BotMenuResponse> {
  return request<BotMenuResponse>("/bot-menu", {
    method: "PUT",
    body: { role, items },
  });
}

export function resetBotMenu(role: BotMenuRole): Promise<BotMenuResponse> {
  return request<BotMenuResponse>(`/bot-menu/reset?role=${role}`, { method: "POST" });
}
