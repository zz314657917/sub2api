import type { CreateOrderResult } from "./payment";

export type StoreProductKind = "card" | "account" | "file";
export type StoreDeliveryStatus =
  "reserved" | "delivered" | "needs_attention" | "released";

export interface StoreProduct {
  id: number;
  name: string;
  description: string;
  kind: StoreProductKind;
  price_cents: number;
  stock_available: number;
  enabled: boolean;
  created_at: string;
  file_id?: number;
}

export interface StoreOrder {
  order_id: number;
  product_id?: number;
  product_name: string;
  kind: StoreProductKind;
  price_cents: number;
  payment_status: string;
  delivery_status: StoreDeliveryStatus;
  created_at: string;
}

export interface StoreDelivery {
  kind: StoreProductKind;
  content?: string;
  username?: string;
  password?: string;
  notes?: string;
  filename?: string;
  download_url?: string;
}

export interface StoreListResponse<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
}

export interface CreateStoreOrderRequest {
  product_id: number;
  idempotency_key: string;
  payment_type: string;
  is_mobile?: boolean;
  payment_source?: string;
  return_url?: string;
  wechat_resume_token?: string;
}

export type CreateStoreOrderResult = CreateOrderResult;

export interface AdminStoreStockItem {
  content?: string;
  username?: string;
  password?: string;
  notes?: string;
}

// 管理端库存列表只返回状态元数据，绝不能在此处增加库存明文内容。
export interface AdminStoreStockRow {
  id: number;
  kind: StoreProductKind;
  state: string;
  created_at: string;
}

export interface AdminStoreOrder extends StoreOrder {
  user_id?: number;
  product_id?: number;
}

export interface StoreFileUpload {
  id: number;
  filename: string;
  size: number;
}
