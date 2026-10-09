import { apiClient } from "./client";
import type {
  AdminStoreOrder,
  AdminStoreStockRow,
  AdminStoreStockItem,
  CreateStoreOrderRequest,
  CreateStoreOrderResult,
  StoreDelivery,
  StoreFileUpload,
  StoreListResponse,
  StoreOrder,
  StoreProduct,
} from "@/types/digitalStore";

export const digitalStoreAPI = {
  getProducts(params?: {
    search?: string;
    kind?: string;
    page?: number;
    page_size?: number;
  }) {
    return apiClient.get<StoreListResponse<StoreProduct>>("/store/products", {
      params,
    });
  },
  createOrder(data: CreateStoreOrderRequest) {
    return apiClient.post<CreateStoreOrderResult>("/store/orders", data);
  },
  getOrders(params?: { page?: number; page_size?: number }) {
    return apiClient.get<StoreListResponse<StoreOrder>>("/store/orders", {
      params,
    });
  },
  getOrder(id: number) {
    return apiClient.get<StoreOrder>(`/store/orders/${id}`);
  },
  getDelivery(id: number) {
    return apiClient.get<StoreDelivery>(`/store/orders/${id}/delivery`);
  },
  resumeOrder(id: number) {
    return apiClient.post<CreateStoreOrderResult>(`/store/orders/${id}/resume`);
  },
  download(id: number) {
    return apiClient.get<Blob>(`/store/orders/${id}/download`, {
      responseType: "blob",
    });
  },
  adminProducts(params?: { page?: number; page_size?: number }) {
    return apiClient.get<StoreListResponse<StoreProduct>>(
      "/admin/store/products",
      { params },
    );
  },
  createProduct(
    data: Omit<StoreProduct, "id" | "created_at" | "stock_available">,
  ) {
    return apiClient.post<StoreProduct>("/admin/store/products", data);
  },
  updateProduct(
    id: number,
    data: Partial<
      Omit<StoreProduct, "id" | "created_at" | "stock_available" | "kind">
    >,
  ) {
    return apiClient.put<StoreProduct>(`/admin/store/products/${id}`, data);
  },
  uploadFile(file: File) {
    const body = new FormData();
    body.append("file", file);
    return apiClient.post<StoreFileUpload>("/admin/store/files", body, {
      headers: { "Content-Type": "multipart/form-data" },
    });
  },
  getStock(id: number, params?: { page?: number; page_size?: number }) {
    return apiClient.get<StoreListResponse<AdminStoreStockRow>>(
      `/admin/store/products/${id}/stock`,
      { params },
    );
  },
  importStock(id: number, items: AdminStoreStockItem[]) {
    return apiClient.post(`/admin/store/products/${id}/stock`, { items });
  },
  adminOrders(params?: {
    page?: number;
    page_size?: number;
    status?: string;
    user_id?: number;
    product_id?: number;
  }) {
    return apiClient.get<StoreListResponse<AdminStoreOrder>>(
      "/admin/store/orders",
      { params },
    );
  },
  retryOrder(id: number) {
    return apiClient.post<AdminStoreOrder>(`/admin/store/orders/${id}/retry`);
  },
};
