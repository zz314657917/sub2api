import { flushPromises, mount } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import StoreManagementView from "@/views/admin/StoreManagementView.vue";

const api = vi.hoisted(() => ({
  adminProducts: vi.fn(),
  adminOrders: vi.fn(),
  getStock: vi.fn(),
  createProduct: vi.fn(),
  updateProduct: vi.fn(),
  uploadFile: vi.fn(),
  importStock: vi.fn(),
  retryOrder: vi.fn(),
}));
vi.mock("@/api/digitalStore", () => ({ digitalStoreAPI: api }));
const product = {
  id: 7,
  name: "测试商品",
  description: "描述",
  kind: "card" as const,
  price_cents: 100,
  stock_available: 2,
  enabled: true,
  created_at: "2026-09-17T00:00:00Z",
};
function render() {
  return mount(StoreManagementView, {
    global: {
      stubs: {
        AppLayout: { template: "<div><slot /></div>" },
        BaseDialog: {
          props: ["show", "title"],
          emits: ["close"],
          template: '<div v-if="show"><slot /><slot name="footer" /></div>',
        },
      },
    },
  });
}
function button(wrapper: ReturnType<typeof render>, label: string) {
  return wrapper.get(`button[aria-label="${label}"]`);
}
beforeEach(() => {
  vi.resetAllMocks();
  api.adminProducts.mockResolvedValue({
    data: { items: [product], total: 21, page: 1, page_size: 20 },
  });
  api.adminOrders.mockResolvedValue({
    data: { items: [], total: 21, page: 1, page_size: 20 },
  });
  api.getStock.mockResolvedValue({
    data: {
      items: [
        { id: 8, kind: "card", state: "available", created_at: "2026-09-17" },
      ],
      total: 21,
      page: 1,
      page_size: 20,
    },
  });
  api.createProduct.mockResolvedValue({ data: product });
  api.updateProduct.mockResolvedValue({ data: product });
  api.uploadFile.mockResolvedValue({
    data: { id: 11, filename: "asset.zip", size: 1 },
  });
  api.importStock.mockResolvedValue({ data: { imported: 1 } });
  api.retryOrder.mockResolvedValue({ data: {} });
});
describe("数字商店后台交互", () => {
  it("保存请求未完成时防止重复提交", async () => {
    let finish!: (value: unknown) => void;
    api.createProduct.mockImplementation(
      () =>
        new Promise((resolve) => {
          finish = resolve;
        }),
    );
    const wrapper = render();
    await flushPromises();
    await wrapper.get("button.btn-primary").trigger("click");
    await wrapper
      .findAll("button")
      .find((item) => item.text() === "保存")!
      .trigger("click");
    await wrapper
      .findAll("button")
      .find((item) => item.text() === "保存中...")!
      .trigger("click");
    expect(api.createProduct).toHaveBeenCalledTimes(1);
    finish({ data: product });
    await flushPromises();
    wrapper.unmount();
  });
  it("在客户端拒绝超过 20 MiB 的上传", async () => {
    const wrapper = render();
    await flushPromises();
    await wrapper.get("button.btn-primary").trigger("click");
    await wrapper.findAll("select").at(-1)!.setValue("file");
    await flushPromises();
    const input = wrapper.get('input[type="file"]');
    Object.defineProperty(input.element, "files", {
      value: [{ size: 20 * 1024 * 1024 + 1, name: "too-large.zip" }],
    });
    await input.trigger("change");
    expect(api.uploadFile).not.toHaveBeenCalled();
    expect(wrapper.get('[role="alert"]').text()).toContain("20 MiB");
    wrapper.unmount();
  });
  it("上传请求失败时显示错误并保持文件商品未保存", async () => {
    api.uploadFile.mockRejectedValue(new Error("upload failed"));
    const wrapper = render();
    await flushPromises();
    await wrapper.get("button.btn-primary").trigger("click");
    await wrapper.findAll("select").at(-1)!.setValue("file");
    await flushPromises();
    const input = wrapper.get('input[type="file"]');
    Object.defineProperty(input.element, "files", {
      value: [{ size: 1, name: "asset.zip" }],
    });
    await input.trigger("change");
    await flushPromises();
    expect(api.uploadFile).toHaveBeenCalledTimes(1);
    expect(wrapper.get('[role="alert"]').text()).toContain("upload failed");
    await wrapper
      .findAll("button")
      .find((item) => item.text() === "保存")!
      .trigger("click");
    expect(api.createProduct).not.toHaveBeenCalled();
    wrapper.unmount();
  });
  it("账号库存保留密码空格且导入中防止重复提交", async () => {
    const accountProduct = { ...product, kind: "account" as const };
    let finish!: (value: unknown) => void;
    api.adminProducts.mockResolvedValue({
      data: { items: [accountProduct], total: 1, page: 1, page_size: 20 },
    });
    api.importStock.mockImplementation(
      () =>
        new Promise((resolve) => {
          finish = resolve;
        }),
    );
    const wrapper = render();
    await flushPromises();
    await wrapper
      .findAll("button")
      .find((item) => item.text() === "管理库存")!
      .trigger("click");
    await flushPromises();
    await wrapper.get("textarea").setValue("alice|  keep spaces  |note");
    await wrapper
      .findAll("button")
      .find((item) => item.text() === "导入")!
      .trigger("click");
    await wrapper
      .findAll("button")
      .find((item) => item.text() === "导入中...")!
      .trigger("click");
    expect(api.importStock).toHaveBeenCalledTimes(1);
    expect(api.importStock).toHaveBeenCalledWith(7, [
      { username: "alice", password: "  keep spaces  ", notes: "note" },
    ]);
    finish({ data: { imported: 1 } });
    await flushPromises();
    wrapper.unmount();
  });
  it("商品、订单和脱敏库存分页都会传递真实页码", async () => {
    const wrapper = render();
    await flushPromises();
    await button(wrapper, "商品下一页").trigger("click");
    await flushPromises();
    expect(api.adminProducts).toHaveBeenLastCalledWith({
      page: 2,
      page_size: 20,
    });
    await button(wrapper, "订单下一页").trigger("click");
    await flushPromises();
    expect(api.adminOrders).toHaveBeenLastCalledWith({
      page: 2,
      page_size: 20,
    });
    await wrapper
      .findAll("button")
      .find((item) => item.text() === "管理库存")!
      .trigger("click");
    await flushPromises();
    await button(wrapper, "库存下一页").trigger("click");
    await flushPromises();
    expect(api.getStock).toHaveBeenLastCalledWith(7, {
      page: 2,
      page_size: 20,
    });
    expect(wrapper.text()).toContain("available");
    wrapper.unmount();
  });
});
