import { beforeEach, describe, expect, it, vi } from "vitest";
import { defineComponent, h, nextTick } from "vue";
import { flushPromises, mount } from "@vue/test-utils";

import EmailTemplateEditor from "../EmailTemplateEditor.vue";

const {
  getEmailTemplates,
  getEmailTemplate,
  updateEmailTemplate,
  restoreOfficialEmailTemplate,
  previewEmailTemplate,
  showError,
  showSuccess,
  routeGuards,
} = vi.hoisted(() => ({
  getEmailTemplates: vi.fn(),
  getEmailTemplate: vi.fn(),
  updateEmailTemplate: vi.fn(),
  restoreOfficialEmailTemplate: vi.fn(),
  previewEmailTemplate: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
  routeGuards: {
    leave: null as null | (() => unknown),
    update: null as null | (() => unknown),
  },
}));

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
  extractApiErrorMessage: () => "api error",
}));

vi.mock("vue-router", async () => {
  const actual = await vi.importActual<typeof import("vue-router")>("vue-router");
  return {
    ...actual,
    onBeforeRouteLeave: (guard: () => unknown) => {
      routeGuards.leave = guard;
    },
    onBeforeRouteUpdate: (guard: () => unknown) => {
      routeGuards.update = guard;
    },
  };
});

vi.mock("vue-i18n", async () => {
  const actual = await vi.importActual<typeof import("vue-i18n")>("vue-i18n");
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
      locale: { value: "zh-CN" },
    }),
  };
});

const ConfirmDialogStub = defineComponent({
  name: "ConfirmDialog",
  props: {
    show: Boolean,
    title: { type: String, default: "" },
    message: { type: String, default: "" },
    confirmText: { type: String, default: "" },
  },
  emits: ["confirm", "cancel"],
  setup(props, { emit }) {
    return () =>
      props.show
        ? h("div", { "data-testid": "confirm-dialog", role: "dialog" }, [
            h("h2", props.title),
            h("p", props.message),
            h(
              "button",
              {
                type: "button",
                "data-testid": "confirm-accept",
                onClick: () => emit("confirm"),
              },
              props.confirmText,
            ),
            h(
              "button",
              {
                type: "button",
                "data-testid": "confirm-cancel",
                onClick: () => emit("cancel"),
              },
              "cancel",
            ),
          ])
        : null;
  },
});

const templateDetails = {
  "auth.verify_code": {
    event: "auth.verify_code",
    locale: "zh-CN",
    subject: "Verification subject",
    html: "<p>Verification body</p>",
    is_custom: false,
    placeholders: ["verification_code"],
  },
  "auth.password_reset": {
    event: "auth.password_reset",
    locale: "zh-CN",
    subject: "Reset subject",
    html: "<p>Reset body</p>",
    is_custom: true,
    placeholders: ["reset_url"],
  },
  "balance.low": {
    event: "balance.low",
    locale: "zh-CN",
    subject: "Balance subject",
    html: "<p>Balance body</p>",
    is_custom: false,
    placeholders: ["current_balance"],
  },
};

function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (reason?: unknown) => void;
  const promise = new Promise<T>((resolvePromise, rejectPromise) => {
    resolve = resolvePromise;
    reject = rejectPromise;
  });
  return { promise, resolve, reject };
}

function mountEditor() {
  return mount(EmailTemplateEditor, {
    global: {
      stubs: {
        ConfirmDialog: ConfirmDialogStub,
      },
    },
  });
}

describe("EmailTemplateEditor", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    routeGuards.leave = null;
    routeGuards.update = null;

    getEmailTemplates.mockResolvedValue({
      events: ["auth.verify_code", "auth.password_reset", "balance.low"],
      locales: ["zh-CN", "en"],
      placeholders: ["site_name"],
    });
    getEmailTemplate.mockImplementation(
      async (event: keyof typeof templateDetails) => templateDetails[event],
    );
    updateEmailTemplate.mockImplementation(
      async (event: string, locale: string, template: { subject: string; html: string }) => ({
        event,
        locale,
        ...template,
        is_custom: true,
      }),
    );
    restoreOfficialEmailTemplate.mockImplementation(
      async (event: string, locale: string) => ({
        event,
        locale,
        subject: "Official subject",
        html: "<p>Official body</p>",
        is_custom: false,
      }),
    );
    previewEmailTemplate.mockImplementation(
      async (request: { subject: string; html: string }) => ({
        subject: `Preview: ${request.subject}`,
        html: `<main>${request.html}</main>`,
      }),
    );
  });

  it("tracks edits against the loaded template and clears dirty state after save", async () => {
    const wrapper = mountEditor();
    await flushPromises();

    expect(wrapper.find("#email-template-unsaved-status").exists()).toBe(false);

    await wrapper.get("#email-template-subject").setValue("Edited subject");
    expect(wrapper.get("#email-template-unsaved-status").attributes("role")).toBe("status");

    await wrapper.get('[data-testid="email-template-save"]').trigger("click");
    await flushPromises();

    expect(updateEmailTemplate).toHaveBeenCalledWith(
      "auth.verify_code",
      "zh-CN",
      {
        subject: "Edited subject",
        html: "<p>Verification body</p>",
      },
    );
    expect(wrapper.find("#email-template-unsaved-status").exists()).toBe(false);
    expect(showSuccess).toHaveBeenCalledWith(
      "admin.settings.emailTemplates.saveSuccess",
    );
  });

  it("keeps the current template when a dirty switch is cancelled and loads the next one after discard", async () => {
    const wrapper = mountEditor();
    await flushPromises();

    await wrapper.get("#email-template-html").setValue("<p>Unsaved</p>");
    await wrapper.get("#email-template-event").setValue("auth.password_reset");
    await nextTick();

    expect(wrapper.get('[data-testid="confirm-dialog"]').text()).toContain(
      "admin.settings.emailTemplates.discardConfirm",
    );
    await wrapper.get('[data-testid="confirm-cancel"]').trigger("click");
    await flushPromises();

    expect((wrapper.get("#email-template-event").element as HTMLSelectElement).value).toBe(
      "auth.verify_code",
    );
    expect(getEmailTemplate).toHaveBeenCalledTimes(1);
    expect((wrapper.get("#email-template-html").element as HTMLTextAreaElement).value).toBe(
      "<p>Unsaved</p>",
    );

    await wrapper.get("#email-template-event").setValue("auth.password_reset");
    await wrapper.get('[data-testid="confirm-accept"]').trigger("click");
    await flushPromises();

    expect(getEmailTemplate).toHaveBeenLastCalledWith(
      "auth.password_reset",
      "zh-CN",
    );
    expect((wrapper.get("#email-template-subject").element as HTMLInputElement).value).toBe(
      "Reset subject",
    );
    expect(wrapper.find("#email-template-unsaved-status").exists()).toBe(false);
  });

  it("uses ConfirmDialog for restoring the official template", async () => {
    const nativeConfirm = vi.spyOn(window, "confirm");
    const wrapper = mountEditor();
    await flushPromises();

    await wrapper.get("#email-template-subject").setValue("Unsaved subject");
    await wrapper.get('[data-testid="email-template-restore"]').trigger("click");

    expect(nativeConfirm).not.toHaveBeenCalled();
    expect(wrapper.get('[data-testid="confirm-dialog"]').text()).toContain(
      "admin.settings.emailTemplates.restoreConfirm",
    );

    await wrapper.get('[data-testid="confirm-accept"]').trigger("click");
    await flushPromises();

    expect(restoreOfficialEmailTemplate).toHaveBeenCalledWith(
      "auth.verify_code",
      "zh-CN",
    );
    expect((wrapper.get("#email-template-subject").element as HTMLInputElement).value).toBe(
      "Official subject",
    );
    expect(wrapper.find("#email-template-unsaved-status").exists()).toBe(false);
    nativeConfirm.mockRestore();
  });

  it("blocks route leave with ConfirmDialog until the user decides", async () => {
    const wrapper = mountEditor();
    await flushPromises();
    await wrapper.get("#email-template-subject").setValue("Unsaved subject");

    expect(routeGuards.leave).not.toBeNull();
    const cancelledNavigation = routeGuards.leave?.() as Promise<boolean>;
    await nextTick();
    await wrapper.get('[data-testid="confirm-cancel"]').trigger("click");
    await expect(cancelledNavigation).resolves.toBe(false);

    const confirmedNavigation = routeGuards.leave?.() as Promise<boolean>;
    await nextTick();
    await wrapper.get('[data-testid="confirm-accept"]').trigger("click");
    await expect(confirmedNavigation).resolves.toBe(true);
    expect(wrapper.find("#email-template-unsaved-status").exists()).toBe(false);
  });

  it("ignores a template response that arrives after a newer selection", async () => {
    const wrapper = mountEditor();
    await flushPromises();

    const slowTemplate = deferred<(typeof templateDetails)["auth.password_reset"]>();
    const currentTemplate = deferred<(typeof templateDetails)["balance.low"]>();
    getEmailTemplate
      .mockImplementationOnce(() => slowTemplate.promise)
      .mockImplementationOnce(() => currentTemplate.promise);

    await wrapper.get("#email-template-event").setValue("auth.password_reset");
    await nextTick();

    const eventSelect = wrapper.get("#email-template-event");
    (eventSelect.element as HTMLSelectElement).value = "balance.low";
    eventSelect.element.dispatchEvent(new Event("change"));
    await nextTick();

    currentTemplate.resolve(templateDetails["balance.low"]);
    await flushPromises();
    slowTemplate.resolve(templateDetails["auth.password_reset"]);
    await flushPromises();

    expect((wrapper.get("#email-template-event").element as HTMLSelectElement).value).toBe(
      "balance.low",
    );
    expect((wrapper.get("#email-template-subject").element as HTMLInputElement).value).toBe(
      "Balance subject",
    );
  });

  it("does not apply a preview response for content edited while previewing", async () => {
    const wrapper = mountEditor();
    await flushPromises();
    const initialPreview = wrapper.get("iframe").attributes("srcdoc");

    const stalePreview = deferred<{ subject: string; html: string }>();
    previewEmailTemplate.mockImplementationOnce(() => stalePreview.promise);

    await wrapper.get('[data-testid="email-template-preview"]').trigger("click");
    await wrapper.get("#email-template-subject").setValue("Newer subject");
    stalePreview.resolve({
      subject: "Stale preview",
      html: "<p>Stale preview</p>",
    });
    await flushPromises();

    expect(wrapper.get("iframe").attributes("srcdoc")).toBe(initialPreview);
  });
});
