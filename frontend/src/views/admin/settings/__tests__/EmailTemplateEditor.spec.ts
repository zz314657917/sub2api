import { flushPromises, mount } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";

import EmailTemplateEditor from "../EmailTemplateEditor.vue";

const {
  getEmailTemplates,
  getEmailTemplate,
  updateEmailTemplate,
  restoreOfficialEmailTemplate,
  previewEmailTemplate,
  showError,
  showSuccess,
} = vi.hoisted(() => ({
  getEmailTemplates: vi.fn(),
  getEmailTemplate: vi.fn(),
  updateEmailTemplate: vi.fn(),
  restoreOfficialEmailTemplate: vi.fn(),
  previewEmailTemplate: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
}));

const localeRef = vi.hoisted(() => ({ value: "zh-CN" }));

vi.mock("@/api", () => ({
  adminAPI: {
    settings: {
      getEmailTemplates,
      getEmailTemplate,
      updateEmailTemplate,
      restoreOfficialEmailTemplate,
      previewEmailTemplate,
    },
  },
}));

vi.mock("@/stores", () => ({
  useAppStore: () => ({ showError, showSuccess }),
}));

vi.mock("@/utils/apiError", () => ({
  extractApiErrorMessage: () => "request failed",
}));

vi.mock("vue-i18n", () => ({
  useI18n: () => ({
    t: (key: string) => key,
    locale: localeRef,
  }),
}));

const template = {
  event: "auth.verify_code",
  locale: "zh",
  subject: "验证码 {{verification_code}}",
  html: "<p>{{verification_code}}</p>",
  placeholders: ["verification_code"],
};

describe("EmailTemplateEditor", () => {
  beforeEach(() => {
    localeRef.value = "zh-CN";
    getEmailTemplates.mockReset();
    getEmailTemplate.mockReset();
    updateEmailTemplate.mockReset();
    restoreOfficialEmailTemplate.mockReset();
    previewEmailTemplate.mockReset();
    showError.mockReset();
    showSuccess.mockReset();

    getEmailTemplates.mockResolvedValue({
      events: [{ value: "auth.verify_code", label: "Verification code" }],
      locales: ["en", "zh"],
      placeholders: ["verification_code"],
    });
    getEmailTemplate.mockResolvedValue(template);
    previewEmailTemplate.mockResolvedValue({
      subject: template.subject,
      html: template.html,
    });
    updateEmailTemplate.mockResolvedValue({ ...template, is_custom: true });
    restoreOfficialEmailTemplate.mockResolvedValue(template);
  });

  it("uses the current console language for the initial template locale", async () => {
    mount(EmailTemplateEditor);
    await flushPromises();

    expect(getEmailTemplate).toHaveBeenCalledWith("auth.verify_code", "zh");
    expect(previewEmailTemplate).toHaveBeenCalledWith(
      expect.objectContaining({ event: "auth.verify_code", locale: "zh" }),
    );
  });

  it("shows only the selected event placeholders when the template API provides them", async () => {
    getEmailTemplates.mockResolvedValue({
      events: [{ value: "auth.verify_code", label: "Verification code" }],
      locales: ["en", "zh"],
      placeholders: ["verification_code", "report_html"],
    });

    const wrapper = mount(EmailTemplateEditor);
    await flushPromises();

    expect(wrapper.text()).toContain("{{verification_code}}");
    expect(wrapper.text()).not.toContain("{{report_html}}");
  });

  it("saves and previews the edited event/locale template", async () => {
    const wrapper = mount(EmailTemplateEditor);
    await flushPromises();

    await wrapper.get("#email-template-subject").setValue("新的主题");
    await wrapper.get("#email-template-html").setValue("<p>新的内容</p>");
    const saveButton = wrapper
      .findAll("button")
      .find((button) => button.text() === "admin.settings.emailTemplates.save");
    expect(saveButton).toBeDefined();
    await saveButton!.trigger("click");
    await flushPromises();

    expect(updateEmailTemplate).toHaveBeenCalledWith("auth.verify_code", "zh", {
      subject: "新的主题",
      html: "<p>新的内容</p>",
    });
    expect(showSuccess).toHaveBeenCalledWith(
      "admin.settings.emailTemplates.saveSuccess",
    );
  });
});
