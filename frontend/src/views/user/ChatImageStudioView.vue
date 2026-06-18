<template>
  <AppLayout>
    <div
      class="chat-image-studio"
      :class="{ 'chat-image-studio-rail-open': railOpen }"
      data-testid="chat-image-studio-view"
    >
      <button
        v-if="railOpen"
        type="button"
        class="studio-mobile-backdrop"
        :aria-label="t('chatImageStudio.closeSessions')"
        @click="railOpen = false"
      ></button>

      <aside class="studio-rail" :class="{ 'studio-rail-mobile-open': railOpen }">
        <div class="studio-settings">
          <div class="studio-rail-mobile-head">
            <span class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('chatImageStudio.sessions') }}</span>
            <button
              type="button"
              class="btn btn-secondary btn-icon lg:hidden"
              :title="t('chatImageStudio.closeSessions')"
              @click="railOpen = false"
            >
              <Icon name="x" size="sm" />
            </button>
          </div>
        </div>

        <div class="studio-rail-actions">
          <div class="studio-key-control">
            <Select
              v-model="selectedKeyId"
              :options="apiKeyOptions"
              :placeholder="loadingKeys ? t('chatImageStudio.loadingKeys') : t('chatImageStudio.selectKey')"
              :disabled="loadingKeys || apiKeyOptions.length === 0"
              searchable="auto"
              data-testid="studio-api-key-select"
            />
          </div>
          <button
            type="button"
            class="btn btn-primary btn-sm w-full justify-center"
            data-testid="studio-new-chat-button"
            @click="startNewSession"
          >
            <Icon name="plus" size="sm" />
            <span>{{ t('chatImageStudio.newChat') }}</span>
          </button>
        </div>

        <div class="studio-session-list custom-scrollbar">
          <article
            v-for="session in sessions"
            :key="session.id"
            class="studio-session-item"
            :class="{ 'studio-session-item-active': session.id === currentSessionId }"
            data-testid="studio-session-item"
          >
            <button type="button" class="studio-session-select" @click="selectSession(session.id)">
              <Icon name="chatBubble" size="sm" class="mt-0.5 flex-shrink-0" />
              <span class="min-w-0 flex-1">
                <span class="block truncate text-sm font-medium">{{ session.title }}</span>
                <span class="mt-0.5 block truncate text-xs text-gray-500 dark:text-dark-300">
                  {{ sessionPreview(session) }}
                </span>
              </span>
            </button>
            <button
              type="button"
              class="studio-session-delete"
              :title="t('chatImageStudio.deleteChat')"
              :aria-label="t('chatImageStudio.deleteChat')"
              :disabled="busy"
              data-testid="studio-delete-session"
              @click="deleteSession(session.id)"
            >
              <Icon name="trash" size="sm" />
            </button>
          </article>
        </div>
      </aside>

      <main class="studio-main">
        <header class="studio-topbar">
          <div class="studio-topbar-left">
            <button
              type="button"
              class="btn btn-secondary btn-icon lg:hidden"
              :title="t('chatImageStudio.sessions')"
              @click="railOpen = !railOpen"
            >
              <Icon name="menu" size="md" />
            </button>
          </div>

          <div class="studio-tabs" role="tablist" :aria-label="t('chatImageStudio.tabsLabel')">
            <button
              v-for="tab in studioTabs"
              :key="tab.value"
              type="button"
              class="studio-tab"
              :class="{ 'studio-tab-active': activeTab === tab.value }"
              role="tab"
              :aria-selected="activeTab === tab.value"
              @click="activeTab = tab.value"
            >
              <Icon :name="tab.icon" size="sm" />
              <span>{{ tab.label }}</span>
            </button>
          </div>

          <div class="studio-status">
            <button
              type="button"
              class="studio-queue-button"
              :class="{ 'studio-queue-button-active': queueOpen }"
              :aria-pressed="queueOpen"
              data-testid="studio-queue-button"
              @click="queueOpen = true"
            >
              <Icon name="clock" size="xs" />
              <span class="studio-queue-label">{{ t('chatImageStudio.queue') }}</span>
              <span class="studio-queue-count">{{ imageTasks.length }}</span>
            </button>
          </div>
        </header>

        <section v-if="activeTab === 'studio'" ref="messagesRef" class="studio-messages custom-scrollbar">
          <div v-if="currentMessages.length === 0" class="studio-empty">
            <div class="studio-empty-icon">
              <Icon name="sparkles" size="xl" />
            </div>
            <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('chatImageStudio.emptyTitle') }}</h2>
            <p class="mt-2 max-w-md text-sm leading-6 text-gray-500 dark:text-dark-300">
              {{ t('chatImageStudio.emptyDescription') }}
            </p>
          </div>

          <div v-else class="studio-message-stack">
            <article
              v-for="message in currentMessages"
              :key="message.id"
              class="studio-message"
              :class="[
                message.role === 'user' ? 'studio-message-user' : 'studio-message-assistant',
                `studio-message-kind-${message.kind}`,
              ]"
              :data-testid="message.role === 'assistant' ? 'studio-message-assistant' : 'studio-message-user'"
            >
              <template v-if="message.role === 'assistant'">
                <div class="studio-avatar studio-avatar-assistant">
                  <Icon :name="message.kind === 'image' ? 'image' : 'sparkles'" size="sm" />
                </div>
                <div class="studio-assistant-body">
                  <div class="studio-message-toolbar">
                    <span>{{ message.kind === 'image' ? t('chatImageStudio.imageResult') : t('chatImageStudio.assistant') }}</span>
                    <div class="studio-message-actions">
                      <button
                        v-if="message.content"
                        type="button"
                        class="studio-message-action"
                        :title="t('chatImageStudio.copyReply')"
                        @click="copyReply(message.content)"
                      >
                        <Icon name="copy" size="xs" />
                        <span>{{ t('common.copy') }}</span>
                      </button>
                      <button
                        v-if="message.kind === 'text'"
                        type="button"
                        class="studio-message-action"
                        :disabled="busy"
                        :title="t('chatImageStudio.resend')"
                        data-testid="studio-resend-assistant"
                        @click="resendAssistantMessage(message)"
                      >
                        <Icon name="refresh" size="xs" />
                        <span>{{ t('chatImageStudio.resend') }}</span>
                      </button>
                    </div>
                  </div>

                  <div v-if="message.kind === 'image'" class="studio-image-message">
                    <div class="mb-3 flex flex-wrap items-center gap-2 text-xs text-gray-500 dark:text-dark-300">
                      <span class="studio-status-chip" :class="{ 'studio-status-chip-live': taskIsActiveStatus(message.status) }">
                        {{ imageStatusLabel(message.status) }}
                      </span>
                      <span v-if="taskIsActiveStatus(message.status)" class="studio-status-chip">
                        <Icon name="clock" size="xs" />
                        <span>{{ t('chatImageStudio.elapsedTime', { time: formattedElapsedTime }) }}</span>
                      </span>
                      <span v-if="message.images?.length" class="studio-status-chip">
                        {{ t('chatImageStudio.imageCount', { count: message.images.length }) }}
                      </span>
                    </div>

                    <p v-if="message.content" class="mb-3 whitespace-pre-wrap break-words text-sm leading-6 text-gray-700 dark:text-dark-100">
                      {{ message.content }}
                    </p>

                    <div v-if="message.images?.length" class="studio-image-batchbar" data-testid="studio-result-batchbar">
                      <span>{{ t('chatImageStudio.selectedImages', { count: selectedCountForImages(message.images) }) }}</span>
                      <div class="studio-image-batch-actions">
                        <button type="button" class="studio-text-action" @click="selectImages(message.images)">
                          {{ t('chatImageStudio.selectAllImages') }}
                        </button>
                        <button
                          type="button"
                          class="studio-text-action"
                          :disabled="selectedCountForImages(message.images) === 0"
                          data-testid="studio-download-selected-result"
                          @click="downloadImages(selectedImagesForImages(message.images))"
                        >
                          <Icon name="download" size="xs" />
                          <span>{{ t('chatImageStudio.downloadSelected') }}</span>
                        </button>
                        <button
                          type="button"
                          class="studio-text-action"
                          :disabled="selectedCountForImages(message.images) === 0"
                          @click="clearImageSelection(message.images)"
                        >
                          {{ t('chatImageStudio.clearSelection') }}
                        </button>
                      </div>
                    </div>

                    <div v-if="message.images?.length" class="studio-image-grid">
                      <article
                        v-for="(image, index) in message.images"
                        :key="image.id"
                        class="studio-image-card"
                        :class="{ 'studio-image-card-selected': isImageSelected(image) }"
                      >
                        <button
                          type="button"
                          class="studio-image-select-toggle"
                          :class="{ 'studio-image-select-toggle-active': isImageSelected(image) }"
                          :aria-label="isImageSelected(image) ? t('chatImageStudio.deselectImage') : t('chatImageStudio.selectImage')"
                          data-testid="studio-image-select"
                          @click.stop="toggleImageSelection(image)"
                        >
                          <Icon v-if="isImageSelected(image)" name="check" size="xs" />
                          <span v-else class="studio-image-select-dot"></span>
                        </button>
                        <button
                          type="button"
                          class="studio-image-preview"
                          :aria-label="t('chatImageStudio.previewImage')"
                          @click="openPreview(image, message.images)"
                        >
                          <img :src="imageSrc(image)" alt="" />
                        </button>
                        <div class="studio-image-card-footer">
                          <span>{{ String(image.outputFormat || 'image').toUpperCase() }}</span>
                          <button type="button" class="studio-icon-action" :title="t('chatImageStudio.download')" @click="downloadImage(image, index)">
                            <Icon name="download" size="sm" />
                          </button>
                        </div>
                      </article>
                    </div>

                    <div v-else-if="taskIsActiveStatus(message.status)" class="studio-generating">
                      <div class="studio-generating-preview">
                        <div class="studio-generating-shine"></div>
                        <div class="studio-generating-icon">
                          <Icon name="sparkles" size="lg" class="animate-pulse" />
                        </div>
                      </div>
                      <div class="min-w-0">
                        <p class="font-medium text-gray-900 dark:text-white">{{ t('chatImageStudio.generatingTitle') }}</p>
                        <p class="mt-1 text-xs text-gray-500 dark:text-dark-300">{{ waitingStepText }}</p>
                        <span class="mt-2 inline-flex">
                          <span class="studio-status-chip">
                            <Icon name="clock" size="xs" />
                            <span>{{ t('chatImageStudio.elapsedTime', { time: formattedElapsedTime }) }}</span>
                          </span>
                        </span>
                        <p v-if="elapsedSeconds >= 60" class="mt-2 text-xs text-gray-500 dark:text-dark-300">
                          {{ t('chatImageStudio.queueLaterHint') }}
                        </p>
                      </div>
                    </div>
                  </div>

                  <div v-else class="studio-message-content whitespace-pre-wrap break-words">
                    <template v-if="message.content">{{ message.content }}</template>
                    <span v-else class="studio-typing">{{ t('chatImageStudio.thinking') }}</span>
                  </div>
                </div>
              </template>

              <template v-else>
                <div class="studio-user-bubble">
                  <div class="studio-message-toolbar studio-message-toolbar-user">
                    <span>{{ t('chatImageStudio.you') }}</span>
                    <div class="studio-message-actions">
                      <button
                        type="button"
                        class="studio-message-action"
                        :disabled="busy"
                        :title="t('chatImageStudio.editMessage')"
                        data-testid="studio-edit-message"
                        @click="editUserMessage(message)"
                      >
                        <Icon name="edit" size="xs" />
                      </button>
                      <button
                        type="button"
                        class="studio-message-action"
                        :disabled="busy"
                        :title="t('chatImageStudio.resend')"
                        data-testid="studio-resend-user"
                        @click="resendUserMessage(message)"
                      >
                        <Icon name="refresh" size="xs" />
                      </button>
                    </div>
                  </div>
                  <div class="studio-message-content whitespace-pre-wrap break-words">
                    {{ message.content }}
                  </div>
                </div>
              </template>
            </article>
          </div>
        </section>

        <section v-else-if="activeTab === 'gallery'" class="studio-gallery custom-scrollbar" data-testid="studio-gallery">
          <div class="studio-section-head">
            <div>
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('chatImageStudio.gallery') }}</h2>
            </div>
            <div class="studio-section-actions">
              <div v-if="selectedImageCount > 0" class="studio-gallery-batchbar" data-testid="studio-gallery-batchbar">
                <span>{{ t('chatImageStudio.selectedImages', { count: selectedImageCount }) }}</span>
                <button type="button" class="studio-text-action" @click="downloadSelectedImages">
                  <Icon name="download" size="xs" />
                  <span>{{ t('chatImageStudio.downloadSelected') }}</span>
                </button>
                <button type="button" class="studio-text-action" @click="clearImageSelection()">
                  {{ t('chatImageStudio.clearSelection') }}
                </button>
              </div>
              <button type="button" class="btn btn-secondary btn-sm" :disabled="loadingImages" @click="loadImageTasks">
                <Icon name="refresh" size="sm" />
                <span>{{ t('common.refresh') }}</span>
              </button>
            </div>
          </div>

          <div v-if="galleryImages.length === 0" class="studio-empty studio-empty-compact">
            <div class="studio-empty-icon">
              <Icon name="image" size="xl" />
            </div>
            <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('chatImageStudio.emptyGalleryTitle') }}</h3>
            <p class="mt-2 text-sm text-gray-500 dark:text-dark-300">{{ t('chatImageStudio.emptyGalleryDescription') }}</p>
          </div>

          <div v-else ref="galleryGridRef" class="studio-gallery-grid" :style="galleryGridStyle">
            <div
              v-for="(column, columnIndex) in galleryColumns"
              :key="`gallery-column-${columnIndex}`"
              class="studio-gallery-column"
            >
              <article
                v-for="{ image, index } in column"
                :key="image.id"
                class="studio-gallery-card"
                :class="{ 'studio-image-card-selected': isImageSelected(image) }"
              >
                <button
                  type="button"
                  class="studio-image-select-toggle"
                  :class="{ 'studio-image-select-toggle-active': isImageSelected(image) }"
                  :aria-label="isImageSelected(image) ? t('chatImageStudio.deselectImage') : t('chatImageStudio.selectImage')"
                  data-testid="studio-gallery-image-select"
                  @click.stop="toggleImageSelection(image)"
                >
                  <Icon v-if="isImageSelected(image)" name="check" size="xs" />
                  <span v-else class="studio-image-select-dot"></span>
                </button>
                <button type="button" class="studio-gallery-preview" :aria-label="t('chatImageStudio.previewImage')" @click="openPreview(image, galleryImages)">
                  <img :src="imageSrc(image)" alt="" />
                </button>
                <button type="button" class="studio-gallery-download" :title="t('chatImageStudio.download')" @click.stop="downloadImage(image, index)">
                  <Icon name="download" size="sm" />
                </button>
              </article>
            </div>
          </div>
        </section>

        <footer v-if="activeTab === 'studio'" class="studio-composer">
          <div class="studio-composer-shell">
            <div
              v-if="referencePreviewUrl"
              class="studio-reference-bubble"
              data-testid="studio-reference-bubble"
              :title="referenceImage?.name || t('chatImageStudio.attachedReferenceImage')"
            >
              <img :src="referencePreviewUrl" alt="" />
              <span class="sr-only">{{ referenceImage?.name }}</span>
              <button
                type="button"
                class="studio-reference-remove"
                :aria-label="t('chatImageStudio.removeReferenceImage')"
                :title="t('chatImageStudio.removeReferenceImage')"
                data-testid="studio-reference-remove"
                @click="clearReferenceImage"
              >
                <Icon name="x" size="xs" />
              </button>
            </div>

            <textarea
              v-model="prompt"
              rows="3"
              class="studio-input"
              :placeholder="inputPlaceholder"
              :disabled="busy"
              data-testid="studio-message-input"
              @paste="onComposerPaste"
              @keydown.enter.exact.prevent="submitPrompt"
            ></textarea>

            <div v-if="mode === 'image' && imageParamsOpen" class="studio-image-params-popover" data-testid="studio-image-params-popover">
              <div class="studio-image-params-callout">
                <strong>{{ t('chatImageStudio.officialImageTool') }}</strong>
              </div>
              <div class="studio-image-params-grid">
                <label class="studio-control-field studio-control-field-count">
                  <span>{{ t('chatImageStudio.countField') }}</span>
                  <Select
                    v-model="imageCount"
                    :options="imageCountOptions"
                    :searchable="false"
                    class="studio-count-select"
                    :aria-label="t('chatImageStudio.imageCountInput')"
                  />
                </label>
                <label class="studio-control-field">
                  <span>{{ t('chatImageStudio.sizeField') }}</span>
                  <Select v-model="imageSize" :options="imageSizeOptions" class="studio-control-select studio-control-small" />
                </label>
                <label class="studio-control-field">
                  <span>{{ t('chatImageStudio.qualityField') }}</span>
                  <Select v-model="imageQuality" :options="imageQualityOptions" class="studio-control-select studio-control-small" />
                </label>
                <label class="studio-control-field">
                  <span>{{ t('chatImageStudio.formatField') }}</span>
                  <Select v-model="outputFormat" :options="outputFormatOptions" class="studio-control-select studio-control-small" />
                </label>
              </div>
              <p class="studio-image-params-note">{{ t('chatImageStudio.imageParamsNote') }}</p>
            </div>

            <div class="studio-composer-toolbar">
              <div class="studio-mode-cluster">
                <div class="studio-mode-switch" role="group" :aria-label="t('chatImageStudio.mode')">
                  <button
                    type="button"
                    class="studio-mode-button"
                    :class="{ 'studio-mode-button-active': mode === 'chat' }"
                    @click="mode = 'chat'; imageParamsOpen = false"
                    >
                      <Icon name="chatBubble" size="sm" />
                      <span class="studio-mode-label">{{ t('chatImageStudio.chatMode') }}</span>
                    </button>
                  <button
                    type="button"
                    class="studio-mode-button"
                    :class="{ 'studio-mode-button-active': mode === 'image' }"
                    @click="mode = 'image'"
                    >
                      <Icon name="image" size="sm" />
                      <span class="studio-mode-label">{{ t('chatImageStudio.imageMode') }}</span>
                    </button>
                </div>

                <template v-if="mode === 'chat'">
                  <div class="studio-chat-model-control studio-mode-model-control" data-testid="studio-chat-controls">
                    <Select
                      v-model="chatModel"
                      :options="chatModelOptions"
                      :placeholder="loadingModels ? t('chatImageStudio.loadingModels') : t('chatImageStudio.selectModel')"
                      class="studio-chat-model-select studio-mode-model-select"
                      searchable="auto"
                      creatable
                      :creatable-prefix="t('chatImageStudio.useCustomModel')"
                      data-testid="studio-chat-model-select"
                    >
                      <template #selected>
                        <span class="studio-model-selected">
                          <Icon name="cpu" size="sm" />
                          <span class="studio-model-selected-label">{{ chatModel }}</span>
                        </span>
                      </template>
                    </Select>
                  </div>
                </template>

                <template v-else>
                  <div class="studio-image-controls studio-mode-model-control" data-testid="studio-image-controls">
                    <input ref="referenceInputRef" type="file" class="hidden" accept="image/png,image/jpeg,image/webp" @change="onReferenceImageChange" />
                    <Select v-model="imageModel" :options="imageModelOptions" class="studio-inline-model-select studio-mode-model-select" data-testid="studio-image-model-select">
                      <template #selected>
                        <span class="studio-model-selected">
                          <Icon name="cpu" size="sm" />
                          <span class="studio-model-selected-label">{{ imageModel }}</span>
                        </span>
                      </template>
                    </Select>
                    <button
                      type="button"
                      class="studio-tool-button studio-params-button"
                      :class="{ 'studio-tool-button-active': imageParamsOpen }"
                      :aria-expanded="imageParamsOpen"
                      data-testid="studio-image-params-button"
                      @click="imageParamsOpen = !imageParamsOpen"
                    >
                      <Icon name="filter" size="sm" />
                      <span class="studio-params-label">{{ t('chatImageStudio.imageParams') }}</span>
                    </button>
                    <button
                      type="button"
                      class="studio-tool-button studio-prompt-market-button"
                      :aria-expanded="promptMarketOpen"
                      data-testid="studio-prompt-market-button"
                      @click="openPromptMarket"
                    >
                      <Icon name="book" size="sm" />
                      <span class="studio-params-label">{{ t('chatImageStudio.promptMarket') }}</span>
                    </button>
                  </div>
                </template>
              </div>

              <div class="studio-submit-group">
                <button
                  v-if="mode === 'image'"
                  type="button"
                  class="studio-circle-action"
                  :class="{ 'studio-circle-action-active': !!referenceImage }"
                  :title="referenceImage ? t('chatImageStudio.attachedReferenceImage') : t('chatImageStudio.referenceImage')"
                  data-testid="studio-reference-upload-button"
                  @click="triggerReferenceUpload"
                >
                  <Icon name="image" size="sm" />
                </button>
                <button
                  v-if="sending"
                  type="button"
                  class="studio-circle-action"
                  :title="t('chatImageStudio.stop')"
                  data-testid="studio-stop-button"
                  @click="stopChat"
                >
                  <Icon name="x" size="sm" />
                </button>
                <button
                  v-else
                  type="button"
                  class="studio-send-button"
                  :disabled="!canSubmit"
                  :title="submitLabel"
                  data-testid="studio-submit-button"
                  @click="submitPrompt"
                >
                  <Icon name="arrowUp" size="sm" />
                </button>
              </div>
            </div>
          </div>
        </footer>
      </main>
    </div>

    <div
      v-if="queueOpen"
      class="studio-queue-overlay"
      data-testid="studio-queue-overlay"
      @click.self="queueOpen = false"
    >
      <section class="studio-queue-modal" role="dialog" aria-modal="true" :aria-label="t('chatImageStudio.queue')" data-testid="studio-queue">
        <header class="studio-queue-modal-head">
          <div class="flex min-w-0 items-center gap-3">
            <span class="studio-queue-modal-icon">
              <Icon name="clipboard" size="sm" />
            </span>
            <div class="min-w-0">
              <h2 class="truncate text-base font-semibold text-gray-900 dark:text-white">{{ t('chatImageStudio.queue') }}</h2>
              <p class="mt-0.5 truncate text-xs text-gray-500 dark:text-dark-300">{{ t('chatImageStudio.queueHint') }}</p>
            </div>
          </div>
          <div class="flex flex-shrink-0 items-center gap-2">
            <button type="button" class="btn btn-secondary btn-sm" :disabled="loadingImages" @click="loadImageTasks">
              <Icon name="refresh" size="sm" />
              <span>{{ t('common.refresh') }}</span>
            </button>
            <button
              type="button"
              class="btn btn-secondary btn-icon"
              :aria-label="t('common.close')"
              data-testid="studio-queue-close"
              @click="queueOpen = false"
            >
              <Icon name="x" size="sm" />
            </button>
          </div>
        </header>

        <div class="studio-queue-modal-body custom-scrollbar">
          <div v-if="imageTasks.length === 0" class="studio-queue-empty">
            <div class="studio-empty-icon">
              <Icon name="checkCircle" size="xl" />
            </div>
            <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('chatImageStudio.emptyQueueTitle') }}</h3>
            <p class="mt-2 max-w-xs text-sm leading-6 text-gray-500 dark:text-dark-300">{{ t('chatImageStudio.emptyQueueDescription') }}</p>
            <button type="button" class="btn btn-primary btn-sm mt-4" @click="queueOpen = false; activeTab = 'studio'">
              <Icon name="sparkles" size="sm" />
              <span>{{ t('chatImageStudio.openStudio') }}</span>
            </button>
          </div>

          <div v-else class="studio-task-list">
            <article v-for="task in imageTasks" :key="task.id" class="studio-task-row">
              <div class="min-w-0 flex-1">
                <div class="flex flex-wrap items-center gap-2">
                  <span class="truncate text-sm font-medium text-gray-900 dark:text-white">{{ task.prompt }}</span>
                  <span class="studio-status-chip" :class="{ 'studio-status-chip-live': taskIsActive(task) }">
                    {{ imageStatusLabel(task.status) }}
                  </span>
                </div>
                <p class="mt-1 truncate text-xs text-gray-500 dark:text-dark-300">
                  {{ displayModelLabel(task.model) }} · {{ task.size }} · {{ task.output_format }} · {{ formatDateTime(task.created_at) }}
                </p>
              </div>
              <button
                v-if="task.images?.length"
                type="button"
                class="btn btn-secondary btn-sm"
                @click="openTaskPreview(task)"
              >
                <Icon name="eye" size="sm" />
                <span>{{ t('chatImageStudio.previewImage') }}</span>
              </button>
            </article>
          </div>
        </div>
      </section>
    </div>

  </AppLayout>

  <Teleport to="body">
    <div
      v-if="promptMarketOpen"
      class="studio-prompt-market-overlay"
      data-testid="studio-prompt-market-overlay"
      @click.self="closePromptMarket"
    >
      <section class="studio-prompt-market-panel" role="dialog" aria-modal="true" :aria-label="t('chatImageStudio.promptMarket')">
        <header class="studio-prompt-market-header">
          <div>
            <h2>{{ t('chatImageStudio.promptMarket') }}</h2>
            <p>{{ t('chatImageStudio.promptMarketSubtitle') }}</p>
          </div>
          <button type="button" class="studio-circle-action" :title="t('common.close')" @click="closePromptMarket">
            <Icon name="x" size="sm" />
          </button>
        </header>

        <div class="studio-prompt-market-filters">
          <label class="studio-prompt-market-search">
            <Icon name="search" size="sm" />
            <input
              v-model="promptMarketQuery"
              type="search"
              :placeholder="t('chatImageStudio.promptMarketSearch')"
              data-testid="studio-prompt-market-search"
            />
          </label>
          <Select
            v-model="promptMarketSource"
            :options="promptMarketSourceOptions"
            :searchable="false"
            class="studio-prompt-market-source"
            data-testid="studio-prompt-market-source"
          />
          <button type="button" class="studio-tool-button" :disabled="promptMarketLoading" @click="reloadPromptMarket">
            <Icon name="refresh" size="sm" />
            <span>{{ t('common.refresh') }}</span>
          </button>
        </div>

        <div v-if="promptMarketError" class="studio-prompt-market-error">
          {{ promptMarketError }}
        </div>

        <div v-if="promptMarketLoading" class="studio-prompt-market-empty">
          {{ t('common.loading') }}
        </div>
        <div v-else-if="filteredPromptMarketItems.length === 0" class="studio-prompt-market-empty">
          {{ t('chatImageStudio.promptMarketEmpty') }}
        </div>
        <div v-else class="studio-prompt-market-list">
          <article
            v-for="item in filteredPromptMarketItems"
            :key="promptFavoriteKey(item)"
            class="studio-prompt-market-card"
            data-testid="studio-prompt-market-card"
          >
            <img :src="item.preview" alt="" loading="lazy" />
            <div class="studio-prompt-market-card-body">
              <div class="studio-prompt-market-card-title-row">
                <h3>{{ item.title }}</h3>
                <button
                  type="button"
                  class="studio-prompt-market-favorite"
                  :class="{ 'studio-prompt-market-favorite-active': isPromptFavorited(item) }"
                  :title="isPromptFavorited(item) ? t('chatImageStudio.removeFavorite') : t('chatImageStudio.addFavorite')"
                  data-testid="studio-prompt-market-favorite"
                  @click="togglePromptFavorite(item)"
                >
                  <Icon name="sparkles" size="sm" />
                </button>
              </div>
              <div class="studio-prompt-market-meta">
                <span>{{ item.category }}</span>
                <span>{{ item.sourceLabel }}</span>
                <span v-if="item.isNsfw">{{ t('chatImageStudio.nsfwPrompt') }}</span>
              </div>
              <p>{{ item.prompt }}</p>
              <div class="studio-prompt-market-actions">
                <button type="button" class="studio-tool-button" data-testid="studio-prompt-market-apply" @click="applyPromptMarketItem(item, false)">
                  <Icon name="check" size="sm" />
                  <span>{{ t('chatImageStudio.applyPrompt') }}</span>
                </button>
                <button
                  v-if="item.referenceImageUrls.length > 0"
                  type="button"
                  class="studio-tool-button"
                  data-testid="studio-prompt-market-apply-reference"
                  @click="applyPromptMarketItem(item, true)"
                >
                  <Icon name="image" size="sm" />
                  <span>{{ t('chatImageStudio.applyPromptWithReference') }}</span>
                </button>
              </div>
            </div>
          </article>
        </div>
      </section>
    </div>
  </Teleport>

  <Teleport to="body">
    <div
      v-if="previewImage"
      class="studio-preview-overlay"
      data-testid="studio-image-preview-overlay"
      @click.self="closePreview"
    >
      <div class="studio-preview-toolbar">
        <div class="studio-preview-toolbar-center">
          <span v-if="previewMetaText" class="studio-preview-pill studio-preview-meta" data-testid="studio-image-preview-meta">
            {{ previewMetaText }}
          </span>
          <span class="studio-preview-pill studio-preview-counter" data-testid="studio-image-preview-counter">
            {{ previewCounterText }}
          </span>
          <div class="studio-preview-actions">
            <button
              type="button"
              class="studio-preview-tool"
              :aria-label="t('chatImageStudio.zoomOut')"
              :disabled="previewZoom <= minPreviewZoom"
              data-testid="studio-image-preview-zoom-out"
              @click.stop="zoomPreview(-0.25)"
            >
              <Icon name="zoomOut" size="sm" />
            </button>
            <span class="studio-preview-zoom" data-testid="studio-image-preview-zoom">{{ previewZoomPercent }}</span>
            <button
              type="button"
              class="studio-preview-tool"
              :aria-label="t('chatImageStudio.zoomIn')"
              :disabled="previewZoom >= maxPreviewZoom"
              data-testid="studio-image-preview-zoom-in"
              @click.stop="zoomPreview(0.25)"
            >
              <Icon name="zoomIn" size="sm" />
            </button>
            <button
              type="button"
              class="studio-preview-tool"
              :aria-label="t('chatImageStudio.resetZoom')"
              data-testid="studio-image-preview-reset-zoom"
              @click.stop="resetPreviewZoom"
            >
              <Icon name="refresh" size="sm" />
            </button>
            <button
              type="button"
              class="studio-preview-tool"
              :aria-label="t('chatImageStudio.download')"
              data-testid="studio-image-preview-download"
              @click.stop="downloadPreviewImage"
            >
              <Icon name="download" size="sm" />
            </button>
          </div>
        </div>
        <button
          type="button"
          class="studio-preview-tool studio-preview-close"
          :aria-label="t('chatImageStudio.closePreview')"
          data-testid="studio-image-preview-close"
          @click.stop="closePreview"
        >
          <Icon name="x" size="md" />
        </button>
      </div>

      <button
        v-if="previewImages.length > 1"
        type="button"
        class="studio-preview-nav studio-preview-nav-prev"
        :aria-label="t('chatImageStudio.previousImage')"
        :disabled="!canPreviewPrevious"
        data-testid="studio-image-preview-prev"
        @click.stop="showPreviousPreview"
      >
        <Icon name="chevronLeft" size="lg" />
      </button>
      <button
        v-if="previewImages.length > 1"
        type="button"
        class="studio-preview-nav studio-preview-nav-next"
        :aria-label="t('chatImageStudio.nextImage')"
        :disabled="!canPreviewNext"
        data-testid="studio-image-preview-next"
        @click.stop="showNextPreview"
      >
        <Icon name="chevronRight" size="lg" />
      </button>

      <div class="studio-preview-body">
        <div class="studio-preview-stage">
          <img
            :src="imageSrc(previewImage)"
            alt=""
            :style="{ transform: `scale(${previewZoom})` }"
            data-testid="studio-image-preview-img"
            @load="onPreviewImageLoad"
          />
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Select, { type SelectOption } from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import { keysAPI, userChannelsAPI } from '@/api'
import type { UserAvailableChannel } from '@/api/channels'
import {
  CHAT_STUDIO_DEFAULT_MODEL,
  createChatCompletionStream,
  isAbortError,
  listChatModels,
  type ChatStudioMessage,
  type ChatStudioModel,
  type ChatStudioRole,
} from '@/api/chatStudio'
import {
  createImageTask,
  downloadImageFile,
  getImageTask,
  listImageTasks,
  type ImageCreatorOutputFormat,
  type ImageCreatorStoredImage,
  type ImageCreatorTask,
  type ImageCreatorTaskStatus,
} from '@/api/imageCreator'
import {
  PROMPT_MARKET_SOURCE_OPTIONS,
  createPromptFavorite,
  deletePromptFavorite,
  fetchPromptFavorites,
  fetchPromptMarketPrompts,
  promptFavoriteKey,
  promptFavoriteRecordKey,
  promptFavoriteToBananaPrompt,
  type BananaPrompt,
  type PromptFavorite,
  type PromptMarketSourceId,
} from '@/api/promptMarket'
import { useClipboard } from '@/composables/useClipboard'
import { useAppStore } from '@/stores'
import type { ApiKey } from '@/types'
import { apiKeyChatGroups, apiKeySupportsChat, apiKeySupportsOpenAIImageGeneration, primaryAPIKeyGroupName } from '@/utils/apiKeyCapabilities'
import { displayModelLabel } from '@/utils/modelDisplay'

type StudioMode = 'chat' | 'image'
type StudioTab = 'studio' | 'gallery'
type StudioMessageKind = 'text' | 'image'

interface GalleryColumnItem {
  image: GeneratedImage
  index: number
}

interface GeneratedImage {
  id: string
  url: string
  sourceUrl: string
  revisedPrompt: string
  outputFormat: ImageCreatorOutputFormat | string
  mimeType: string
  byteSize?: number
  width?: number
  height?: number
}

interface DownloadableImage {
  image: GeneratedImage
  index: number
}

interface StudioMessage {
  id: string
  role: ChatStudioRole
  kind: StudioMessageKind
  content: string
  createdAt: string
  taskId?: number
  status?: ImageCreatorTaskStatus
  images?: GeneratedImage[]
}

interface StudioSession {
  id: string
  title: string
  messages: StudioMessage[]
  createdAt: string
  updatedAt: string
}

interface StudioStoragePayload {
  sessions?: StudioSession[]
  currentSessionId?: string | null
}

const STORAGE_KEY = 'sub2api:chat-image-studio:v1'
const MAX_SESSIONS = 20
const MAX_MESSAGES_PER_SESSION = 120
const SESSION_TITLE_LIMIT = 28
const maxImageCount = 8
const minPreviewZoom = 0.25
const maxPreviewZoom = 4
const imagePreviewFallbackSrc = 'data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///ywAAAAAAQABAAACAUwAOw=='

const { t } = useI18n()
const route = useRoute()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()

const activeTab = ref<StudioTab>('studio')
const mode = ref<StudioMode>('image')
const railOpen = ref(false)
const queueOpen = ref(false)
const imageParamsOpen = ref(false)
const prompt = ref('')
const editingMessageId = ref<string | null>(null)
const sessions = ref<StudioSession[]>([])
const currentSessionId = ref<string | null>(null)
const messagesRef = ref<HTMLElement | null>(null)
const galleryGridRef = ref<HTMLElement | null>(null)

const apiKeys = ref<ApiKey[]>([])
const channels = ref<UserAvailableChannel[]>([])
const keyModels = ref<ChatStudioModel[]>([])
const selectedKeyId = ref<number | null>(null)
const chatModel = ref(CHAT_STUDIO_DEFAULT_MODEL)
const loadingKeys = ref(false)
const loadingModels = ref(false)
const loadingImages = ref(false)
const sending = ref(false)
const generating = ref(false)

const imageModel = ref('gpt-image-2')
const imageSize = ref('auto')
const imageQuality = ref('auto')
const outputFormat = ref<ImageCreatorOutputFormat>('webp')
const imageCount = ref(1)
const background = ref('auto')
const referenceImage = ref<File | null>(null)
const referencePreviewUrl = ref('')
const referenceInputRef = ref<HTMLInputElement | null>(null)
const galleryImages = ref<GeneratedImage[]>([])
const imageTasks = ref<ImageCreatorTask[]>([])
const previewImage = ref<GeneratedImage | null>(null)
const previewZoom = ref(1)
const previewImages = ref<GeneratedImage[]>([])
const imageDisplayUrls = ref<Record<string, string>>({})
const selectedImageIds = ref<string[]>([])
const elapsedSeconds = ref(0)
const waitingStepIndex = ref(0)
const activeTaskId = ref<number | null>(null)
const galleryColumnCount = ref(4)
const promptMarketOpen = ref(false)
const promptMarketLoading = ref(false)
const promptMarketError = ref('')
const promptMarketItems = ref<BananaPrompt[]>([])
const promptMarketFavorites = ref<PromptFavorite[]>([])
const promptMarketQuery = ref('')
const promptMarketSource = ref<PromptMarketSourceId | 'all' | 'favorites'>('all')

let abortController: AbortController | null = null
let promptMarketAbortController: AbortController | null = null
let taskPollTimerId: ReturnType<typeof setInterval> | null = null
let generationTimerId: ReturnType<typeof setInterval> | null = null
let activePollMessageId: string | null = null
let loadModelsRequestId = 0

const generatedImageObjectUrls = new Set<string>()

const imageModelOptions = [
  { value: 'gpt-image-2', label: 'gpt-image-2' },
  { value: 'gpt-image-1.5', label: 'gpt-image-1.5' },
  { value: 'gpt-image-1', label: 'gpt-image-1' },
]

const imageSizeOptions = [
  { value: 'auto', label: 'Auto' },
  { value: '1024x1024', label: '1:1 1024x1024' },
  { value: '1536x1024', label: '3:2 1536x1024' },
  { value: '1024x1536', label: '2:3 1024x1536' },
  { value: '2048x2048', label: '1:1 2048x2048' },
  { value: '3840x2160', label: '16:9 3840x2160' },
  { value: '2160x3840', label: '9:16 2160x3840' },
]

const imageCountOptions = Array.from({ length: maxImageCount }, (_, index) => {
  const count = index + 1
  return { value: count, label: String(count) }
})

const imageQualityOptions = [
  { value: 'auto', label: 'auto' },
  { value: 'low', label: 'low' },
  { value: 'medium', label: 'medium' },
  { value: 'high', label: 'high' },
]

const outputFormatOptions = [
  { value: 'webp', label: 'WEBP' },
  { value: 'jpeg', label: 'JPEG' },
  { value: 'png', label: 'PNG' },
]

const waitingStepKeys = [
  'chatImageStudio.waitingSteps.routing',
  'chatImageStudio.waitingSteps.rendering',
  'chatImageStudio.waitingSteps.receiving',
  'chatImageStudio.waitingSteps.finishing',
]

const studioTabs = computed<Array<{ value: StudioTab; label: string; icon: 'chatBubble' | 'image' }>>(() => [
  { value: 'studio', label: t('chatImageStudio.studio'), icon: 'chatBubble' },
  { value: 'gallery', label: t('chatImageStudio.gallery'), icon: 'image' },
])

const promptMarketSourceOptions = computed(() =>
  PROMPT_MARKET_SOURCE_OPTIONS.map((option) => ({
    value: option.value,
    label: option.value === 'all'
      ? t('chatImageStudio.allPrompts')
      : option.value === 'favorites'
        ? t('chatImageStudio.favoritePrompts')
        : option.label,
  }))
)

const currentSession = computed(() =>
  sessions.value.find((session) => session.id === currentSessionId.value) ?? null
)

const currentMessages = computed(() => currentSession.value?.messages ?? [])

const selectedKey = computed(() =>
  apiKeys.value.find((key) => key.id === selectedKeyId.value) ?? null
)

const apiKeyOptions = computed<SelectOption[]>(() =>
  apiKeys.value.map((key) => ({
    value: key.id,
    label: apiKeyLabel(key),
  }))
)

const chatModelOptions = computed<SelectOption[]>(() => {
  const groupIds = new Set(
    (selectedKey.value ? apiKeyChatGroups(selectedKey.value).map((group) => group.id) : [])
      .filter((id): id is number => typeof id === 'number' && id > 0)
  )
  const names = new Set<string>()

  for (const keyModel of keyModels.value) {
    if (keyModel.id) names.add(keyModel.id)
  }

  for (const channel of channels.value) {
    for (const section of channel.platforms || []) {
      if (groupIds.size > 0 && !section.groups?.some((group) => groupIds.has(group.id))) continue
      for (const supportedModel of section.supported_models || []) {
        if (supportedModel.name) names.add(supportedModel.name)
      }
    }
  }

  if (!names.has(chatModel.value)) names.add(chatModel.value || CHAT_STUDIO_DEFAULT_MODEL)
  if (!names.has(CHAT_STUDIO_DEFAULT_MODEL)) names.add(CHAT_STUDIO_DEFAULT_MODEL)

  return Array.from(names)
    .filter(Boolean)
    .sort((a, b) => a.localeCompare(b))
    .map((name) => ({ value: name, label: name }))
})

const busy = computed(() => sending.value || generating.value)

const canSubmit = computed(() => {
  if (busy.value || prompt.value.trim().length === 0) return false
  if (!selectedKey.value?.key) return false
  if (mode.value === 'chat') return chatModel.value.trim().length > 0
  return isOpenAIImageKey(selectedKey.value) && imageCount.value >= 1
})

const submitLabel = computed(() =>
  mode.value === 'image' ? t('chatImageStudio.generate') : t('chatImageStudio.send')
)

const inputPlaceholder = computed(() =>
  mode.value === 'image' ? t('chatImageStudio.imagePlaceholder') : t('chatImageStudio.chatPlaceholder')
)

const waitingStepText = computed(() => t(waitingStepKeys[waitingStepIndex.value]))

const formattedElapsedTime = computed(() => formatDuration(elapsedSeconds.value))

const selectedImageCount = computed(() => selectedImageIds.value.length)

const selectedGalleryImages = computed<DownloadableImage[]>(() =>
  galleryImages.value
    .map((image, index) => ({ image, index }))
    .filter(({ image }) => isImageSelected(image))
)

const promptFavoriteRecords = computed(() => {
  const records = new Map<string, PromptFavorite>()
  for (const favorite of promptMarketFavorites.value) {
    records.set(promptFavoriteRecordKey(favorite), favorite)
  }
  return records
})

const promptMarketVisibleItems = computed<BananaPrompt[]>(() => {
  if (promptMarketSource.value === 'favorites') {
    return promptMarketFavorites.value.map(promptFavoriteToBananaPrompt)
  }
  if (promptMarketSource.value === 'all') {
    return promptMarketItems.value
  }
  return promptMarketItems.value.filter((item) => item.source === promptMarketSource.value)
})

const filteredPromptMarketItems = computed<BananaPrompt[]>(() => {
  const query = promptMarketQuery.value.trim().toLowerCase()
  if (!query) return promptMarketVisibleItems.value
  return promptMarketVisibleItems.value.filter((item) =>
    [
      item.title,
      item.prompt,
      item.category,
      item.subCategory || '',
      item.sourceLabel,
      item.author,
    ].join('\n').toLowerCase().includes(query)
  )
})

const galleryColumns = computed<GalleryColumnItem[][]>(() => {
  const count = Math.max(1, galleryColumnCount.value)
  const columns = Array.from({ length: count }, () => [] as GalleryColumnItem[])
  galleryImages.value.forEach((image, index) => {
    columns[index % count].push({ image, index })
  })
  return columns.filter((column) => column.length > 0)
})

const galleryGridStyle = computed(() => ({
  gridTemplateColumns: `repeat(${Math.max(1, galleryColumns.value.length)}, minmax(0, 220px))`,
}))

const previewZoomPercent = computed(() => `${Math.round(previewZoom.value * 100)}%`)

const previewIndex = computed(() => {
  if (!previewImage.value) return -1
  return previewImages.value.findIndex((image) => image.id === previewImage.value?.id)
})

const canPreviewPrevious = computed(() => previewIndex.value > 0)

const canPreviewNext = computed(() => previewIndex.value >= 0 && previewIndex.value < previewImages.value.length - 1)

const previewCounterText = computed(() => {
  const total = Math.max(1, previewImages.value.length)
  const index = previewIndex.value >= 0 ? previewIndex.value + 1 : 1
  return `${index} / ${total}`
})

const previewMetaText = computed(() => {
  if (!previewImage.value) return ''
  const parts = [
    formatBytes(previewImage.value.byteSize),
    formatImageDimensions(previewImage.value),
    formatImageAspectRatio(previewImage.value),
    formatImageMegapixels(previewImage.value),
  ].filter(Boolean)
  return parts.join(' · ')
})

watch([sessions, currentSessionId], () => {
  persistSessions()
}, { deep: true })

watch(selectedKeyId, () => {
  keyModels.value = []
  void loadModels()
})

onMounted(() => {
  restoreSessions()
  applyRouteDraft()
  void loadApiKeys()
  void loadImageTasks()
  void nextTick(updateGalleryColumnCount)
  window.addEventListener('resize', updateGalleryColumnCount)
  window.addEventListener('keydown', onPreviewKeydown)
})

watch(
  () => [route.query.prompt, route.query.mode, route.query.reference_image_id],
  () => applyRouteDraft(),
)

watch(activeTab, async (tab) => {
  if (tab !== 'gallery') return
  await nextTick()
  updateGalleryColumnCount()
})

watch(galleryImages, async () => {
  if (activeTab.value !== 'gallery') return
  await nextTick()
  updateGalleryColumnCount()
})

onBeforeUnmount(() => {
  abortController?.abort()
  promptMarketAbortController?.abort()
  stopTaskPolling()
  stopGenerationTimer()
  revokeGeneratedImageObjectUrls()
  clearReferenceImage()
  window.removeEventListener('resize', updateGalleryColumnCount)
  window.removeEventListener('keydown', onPreviewKeydown)
})

function createId(prefix: string): string {
  return `${prefix}_${Date.now().toString(36)}_${Math.random().toString(36).slice(2, 8)}`
}

function nowIso(): string {
  return new Date().toISOString()
}

function createEmptySession(title = t('chatImageStudio.defaultSessionTitle')): StudioSession {
  const now = nowIso()
  return {
    id: createId('studio'),
    title,
    messages: [],
    createdAt: now,
    updatedAt: now,
  }
}

function restoreSessions(): void {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) {
      const session = createEmptySession()
      sessions.value = [session]
      currentSessionId.value = session.id
      return
    }

    const payload = JSON.parse(raw) as StudioStoragePayload
    const restored = normalizeSessions(payload.sessions || [])
    if (restored.length === 0) {
      const session = createEmptySession()
      sessions.value = [session]
      currentSessionId.value = session.id
      return
    }

    sessions.value = restored
    currentSessionId.value =
      restored.find((session) => session.id === payload.currentSessionId)?.id ??
      restored[0]?.id ??
      null
    void hydrateImages(imagesFromSessions())
  } catch {
    const session = createEmptySession()
    sessions.value = [session]
    currentSessionId.value = session.id
  }
}

function applyRouteDraft(): void {
  const rawPrompt = Array.isArray(route.query.prompt) ? route.query.prompt[0] : route.query.prompt
  const draft = typeof rawPrompt === 'string' ? rawPrompt.trim() : ''
  if (draft) {
    prompt.value = draft
    activeTab.value = 'studio'
  }
  const rawMode = Array.isArray(route.query.mode) ? route.query.mode[0] : route.query.mode
  if (rawMode === 'chat' || rawMode === 'image') {
    mode.value = rawMode
  } else if (draft) {
    mode.value = 'image'
  }
  const rawReferenceID = Array.isArray(route.query.reference_image_id) ? route.query.reference_image_id[0] : route.query.reference_image_id
  if (typeof rawReferenceID === 'string' && rawReferenceID.trim()) {
    void attachManagedReferenceImage(rawReferenceID.trim())
  }
}

async function attachManagedReferenceImage(rawID: string): Promise<void> {
  const id = Number(rawID)
  if (!Number.isFinite(id) || id <= 0) {
    return
  }
  const url = `/api/v1/user/image-creator/images/${Math.trunc(id)}/reference-file`
  try {
    const blob = await downloadImageFile(url)
    const mimeType = blob.type || 'image/png'
    const ext = referenceImageExtension(mimeType)
    setReferenceImage(new File([blob], `managed-reference-${Math.trunc(id)}.${ext}`, { type: mimeType }))
    mode.value = 'image'
    activeTab.value = 'studio'
    appStore.showSuccess(t('chatImageStudio.referenceAttachedFromLibrary'))
  } catch (error: any) {
    appStore.showError(error?.message || t('chatImageStudio.referenceAttachFailed'))
  }
}

function referenceImageExtension(mimeType: string): string {
  switch (mimeType.toLowerCase()) {
    case 'image/jpeg':
      return 'jpg'
    case 'image/webp':
      return 'webp'
    default:
      return 'png'
  }
}

function normalizeSessions(input: StudioSession[]): StudioSession[] {
  return input
    .filter((session) => session && typeof session.id === 'string')
    .map((session) => ({
      ...session,
      title: String(session.title || t('chatImageStudio.defaultSessionTitle')),
      messages: Array.isArray(session.messages)
        ? session.messages
          .filter((message) =>
            message &&
            ['system', 'user', 'assistant'].includes(message.role) &&
            typeof message.content === 'string'
          )
          .map(normalizeMessage)
          .slice(-MAX_MESSAGES_PER_SESSION)
        : [],
      createdAt: session.createdAt || nowIso(),
      updatedAt: session.updatedAt || nowIso(),
    }))
    .sort((a, b) => Date.parse(b.updatedAt) - Date.parse(a.updatedAt))
    .slice(0, MAX_SESSIONS)
}

function normalizeMessage(message: StudioMessage): StudioMessage {
  const images = Array.isArray(message.images)
    ? message.images.map((image, index) => normalizeGeneratedImage(image, index)).filter(Boolean) as GeneratedImage[]
    : undefined
  return {
    id: typeof message.id === 'string' ? message.id : createId('msg'),
    role: message.role,
    kind: message.kind === 'image' ? 'image' : 'text',
    content: String(message.content || ''),
    createdAt: message.createdAt || nowIso(),
    taskId: typeof message.taskId === 'number' ? message.taskId : undefined,
    status: message.status,
    images,
  }
}

function normalizeGeneratedImage(image: GeneratedImage, index: number): GeneratedImage | null {
  const sourceUrl = String(image.sourceUrl || image.url || '').trim()
  if (!sourceUrl) return null
  return {
    id: String(image.id || `${Date.now()}-${index}`),
    url: sourceUrl,
    sourceUrl,
    revisedPrompt: String(image.revisedPrompt || ''),
    outputFormat: image.outputFormat || 'webp',
    mimeType: image.mimeType || '',
    byteSize: typeof image.byteSize === 'number' && Number.isFinite(image.byteSize) ? image.byteSize : undefined,
    width: typeof image.width === 'number' && Number.isFinite(image.width) ? image.width : undefined,
    height: typeof image.height === 'number' && Number.isFinite(image.height) ? image.height : undefined,
  }
}

function persistSessions(): void {
  const serializable = normalizeSessions(sessions.value).map((session) => ({
    ...session,
    messages: session.messages.map((message) => ({
      ...message,
      images: message.images?.map((image) => ({
        ...image,
        url: image.sourceUrl || image.url,
      })),
    })),
  }))
  localStorage.setItem(STORAGE_KEY, JSON.stringify({
    sessions: serializable,
    currentSessionId: currentSessionId.value,
  }))
}

function touchSession(session: StudioSession): void {
  session.updatedAt = nowIso()
  if (session.messages.length > MAX_MESSAGES_PER_SESSION) {
    session.messages = session.messages.slice(-MAX_MESSAGES_PER_SESSION)
  }
}

function ensureCurrentSession(): StudioSession {
  const existing = currentSession.value
  if (existing) return existing

  const session = createEmptySession()
  sessions.value.unshift(session)
  currentSessionId.value = session.id
  return session
}

function trimSessions(): void {
  const normalized = normalizeSessions(sessions.value)
  const sameOrder =
    normalized.length === sessions.value.length &&
    normalized.every((session, index) => session.id === sessions.value[index]?.id)
  if (!sameOrder) {
    sessions.value = normalized
  }
  if (!sessions.value.some((session) => session.id === currentSessionId.value)) {
    currentSessionId.value = sessions.value[0]?.id ?? null
  }
}

function startNewSession(): void {
  if (sending.value) stopChat()
  const session = createEmptySession()
  sessions.value.unshift(session)
  currentSessionId.value = session.id
  prompt.value = ''
  railOpen.value = false
  activeTab.value = 'studio'
  trimSessions()
}

function selectSession(id: string): void {
  currentSessionId.value = id
  railOpen.value = false
  activeTab.value = 'studio'
  void scrollToBottom()
}

function deleteSession(id: string): void {
  if (busy.value) return
  const index = sessions.value.findIndex((session) => session.id === id)
  if (index < 0) return
  sessions.value.splice(index, 1)
  if (sessions.value.length === 0) {
    const session = createEmptySession()
    sessions.value = [session]
    currentSessionId.value = session.id
    return
  }
  if (id === currentSessionId.value) {
    currentSessionId.value = sessions.value[Math.max(0, index - 1)]?.id ?? sessions.value[0]?.id ?? null
  }
}

function sessionPreview(session: StudioSession): string {
  const last = [...session.messages].reverse().find((message) => message.content.trim() || message.images?.length)
  if (!last) return t('chatImageStudio.emptySession')
  if (last.kind === 'image' && last.images?.length) {
    return t('chatImageStudio.imageCount', { count: last.images.length })
  }
  return last.content.trim() || t('chatImageStudio.emptySession')
}

function apiKeyLabel(key: ApiKey): string {
  return `API· ${key.name || primaryAPIKeyGroupName(key) || key.group?.platform || t('chatImageStudio.apiKey')}`
}

function isUsableKey(key: ApiKey): boolean {
  return !!key.key && apiKeySupportsChat(key)
}

function isOpenAIImageKey(key: ApiKey): boolean {
  return apiKeySupportsOpenAIImageGeneration(key)
}

function pickDefaultKey(keys: ApiKey[]): ApiKey | null {
  const current = keys.find((key) => key.id === selectedKeyId.value)
  return current ?? keys.find(isOpenAIImageKey) ?? keys[0] ?? null
}

async function loadApiKeys(): Promise<void> {
  loadingKeys.value = true
  try {
    const response = await keysAPI.list(1, 100, {
      status: 'active',
      sort_by: 'created_at',
      sort_order: 'desc',
    })
    apiKeys.value = response.items.filter(isUsableKey)
    const previousSelectedKeyId = selectedKeyId.value
    selectedKeyId.value = pickDefaultKey(apiKeys.value)?.id ?? null
    if (selectedKeyId.value === previousSelectedKeyId) {
      void loadModels()
    }
  } catch {
    appStore.showError(t('chatImageStudio.loadKeysFailed'))
  } finally {
    loadingKeys.value = false
  }
}

async function loadModels(): Promise<void> {
  loadingModels.value = true
  const key = selectedKey.value?.key || ''
  const requestId = ++loadModelsRequestId
  try {
    const [models, availableChannels] = await Promise.all([
      key ? listChatModels(key).catch((): ChatStudioModel[] => []) : Promise.resolve<ChatStudioModel[]>([]),
      userChannelsAPI.getAvailable().catch((): UserAvailableChannel[] => []),
    ])
    if (requestId !== loadModelsRequestId) return
    keyModels.value = models
    channels.value = availableChannels
    if (!chatModel.value.trim()) {
      chatModel.value = chatModelOptions.value[0]?.value ? String(chatModelOptions.value[0].value) : CHAT_STUDIO_DEFAULT_MODEL
    }
  } catch {
    if (requestId !== loadModelsRequestId) return
    keyModels.value = []
    channels.value = []
    chatModel.value = chatModel.value.trim() || CHAT_STUDIO_DEFAULT_MODEL
    appStore.showWarning(t('chatImageStudio.loadModelsFailed'))
  } finally {
    if (requestId === loadModelsRequestId) {
      loadingModels.value = false
    }
  }
}

function buildConversationMessages(session: StudioSession): ChatStudioMessage[] {
  return session.messages
    .filter((message) => message.kind === 'text' && message.content.trim().length > 0)
    .map((message) => ({
      role: message.role,
      content: message.content,
    }))
}

function updateTitleFromPrompt(session: StudioSession, text: string): void {
  if (session.title !== t('chatImageStudio.defaultSessionTitle') && session.title.trim()) return
  const normalized = text.replace(/\s+/g, ' ').trim()
  if (!normalized) return
  session.title = normalized.length > SESSION_TITLE_LIMIT
    ? `${normalized.slice(0, SESSION_TITLE_LIMIT)}...`
    : normalized
}

async function submitPrompt(): Promise<void> {
  if (mode.value === 'image') {
    await generateImage()
    return
  }
  await sendChat()
}

async function sendChat(): Promise<void> {
  const text = prompt.value.trim()
  if (!text) {
    appStore.showError(t('chatImageStudio.emptyMessage'))
    return
  }
  if (!selectedKey.value?.key) {
    appStore.showError(t('chatImageStudio.noApiKey'))
    return
  }
  if (!chatModel.value.trim()) {
    appStore.showError(t('chatImageStudio.noModel'))
    return
  }

  const session = ensureCurrentSession()
  prompt.value = ''
  const editingMessage = editingMessageId.value ? findMessageById(editingMessageId.value) : null
  if (editingMessage?.role === 'user') {
    const editSession = sessionForMessage(editingMessage.id)
    if (editSession) {
      trimConversationAfterMessage(editSession, editingMessage.id)
      editingMessageId.value = null
      await appendChatExchange(editSession, text, selectedKey.value.key, chatModel.value.trim())
      return
    }
  }

  editingMessageId.value = null
  updateTitleFromPrompt(session, text)
  await appendChatExchange(session, text, selectedKey.value.key, chatModel.value.trim())
}

async function appendChatExchange(session: StudioSession, text: string, apiKey: string, model: string): Promise<void> {
  const userMessage: StudioMessage = {
    id: createId('msg'),
    role: 'user',
    kind: 'text',
    content: text,
    createdAt: nowIso(),
  }
  const assistantMessage: StudioMessage = {
    id: createId('msg'),
    role: 'assistant',
    kind: 'text',
    content: '',
    createdAt: nowIso(),
  }
  session.messages.push(userMessage, assistantMessage)
  touchSession(session)
  activeTab.value = 'studio'
  void scrollToBottom()

  abortController = new AbortController()
  sending.value = true
  const requestMessages = buildConversationMessages(session)

  try {
    await createChatCompletionStream({
      apiKey,
      model,
      messages: requestMessages,
      signal: abortController.signal,
      onDelta: (delta) => {
        assistantMessage.content += delta
        touchSession(session)
        void scrollToBottom()
      },
    })
    if (!assistantMessage.content.trim()) {
      assistantMessage.content = t('chatImageStudio.emptyAssistantReply')
    }
  } catch (error: unknown) {
    if (isAbortError(error)) {
      if (!assistantMessage.content.trim()) {
        session.messages = session.messages.filter((message) => message.id !== assistantMessage.id)
      }
      return
    }

    const message = error instanceof Error ? error.message : t('chatImageStudio.requestFailed')
    const displayMessage = t('chatImageStudio.requestFailedWithMessage', { message })
    if (!assistantMessage.content.trim()) {
      assistantMessage.content = displayMessage
    }
    appStore.showError(displayMessage)
  } finally {
    touchSession(session)
    sending.value = false
    abortController = null
    void scrollToBottom()
  }
}

function editUserMessage(message: StudioMessage): void {
  if (busy.value || message.role !== 'user') return
  mode.value = 'chat'
  prompt.value = message.content
  editingMessageId.value = message.id
  activeTab.value = 'studio'
}

async function resendUserMessage(message: StudioMessage): Promise<void> {
  if (busy.value || message.role !== 'user' || !message.content.trim()) return
  if (!selectedKey.value?.key) {
    appStore.showError(t('chatImageStudio.noApiKey'))
    return
  }
  if (!chatModel.value.trim()) {
    appStore.showError(t('chatImageStudio.noModel'))
    return
  }

  const session = sessionForMessage(message.id)
  if (!session) return
  trimConversationAfterMessage(session, message.id)
  editingMessageId.value = null
  await appendChatExchange(session, message.content.trim(), selectedKey.value.key, chatModel.value.trim())
}

async function resendAssistantMessage(message: StudioMessage): Promise<void> {
  if (busy.value || message.role !== 'assistant' || message.kind !== 'text') return
  const session = sessionForMessage(message.id)
  if (!session) return
  const messageIndex = session.messages.findIndex((item) => item.id === message.id)
  if (messageIndex <= 0) return
  const previousUserMessage = [...session.messages]
    .slice(0, messageIndex)
    .reverse()
    .find((item) => item.role === 'user' && item.kind === 'text' && item.content.trim())
  if (!previousUserMessage) return
  await resendUserMessage(previousUserMessage)
}

function sessionForMessage(messageId: string): StudioSession | null {
  return sessions.value.find((session) => session.messages.some((message) => message.id === messageId)) ?? null
}

function trimConversationAfterMessage(session: StudioSession, messageId: string): void {
  const index = session.messages.findIndex((message) => message.id === messageId)
  if (index < 0) return
  session.messages = session.messages.slice(0, index)
  currentSessionId.value = session.id
  activeTab.value = 'studio'
  touchSession(session)
}

function stopChat(): void {
  abortController?.abort()
}

async function generateImage(): Promise<void> {
  const text = prompt.value.trim()
  if (!text) {
    appStore.showError(t('chatImageStudio.emptyMessage'))
    return
  }
  if (!selectedKey.value || !isOpenAIImageKey(selectedKey.value)) {
    appStore.showError(t('chatImageStudio.selectImageKeyFirst'))
    return
  }

  const session = ensureCurrentSession()
  prompt.value = ''
  imageCount.value = clampImageCount()
  updateTitleFromPrompt(session, text)

  const userMessage: StudioMessage = {
    id: createId('msg'),
    role: 'user',
    kind: 'text',
    content: text,
    createdAt: nowIso(),
  }
  const assistantMessage: StudioMessage = {
    id: createId('img'),
    role: 'assistant',
    kind: 'image',
    content: t('chatImageStudio.generatingHint'),
    createdAt: nowIso(),
    status: 'pending',
    images: [],
  }
  session.messages.push(userMessage, assistantMessage)
  touchSession(session)
  activeTab.value = 'studio'
  generating.value = true
  startGenerationTimer()
  void scrollToBottom()

  try {
    const task = await createImageTask({
      apiKeyId: selectedKey.value.id,
      model: imageModel.value,
      prompt: text,
      size: imageSize.value,
      quality: imageQuality.value,
      count: imageCount.value,
      outputFormat: outputFormat.value,
      background: background.value,
      referenceImage: referenceImage.value,
    })
    assistantMessage.taskId = task.id
    assistantMessage.status = task.status
    upsertTask(task)
    startTaskPolling(task.id, assistantMessage.id)
  } catch (error: any) {
    stopGenerationTimer()
    generating.value = false
    assistantMessage.status = 'failed'
    assistantMessage.content = error?.message || t('chatImageStudio.generateFailed')
    appStore.showError(assistantMessage.content)
  } finally {
    touchSession(session)
  }
}

function clampImageCount(): number {
  const n = Number(imageCount.value)
  if (!Number.isFinite(n)) return 1
  return Math.min(Math.max(Math.trunc(n), 1), maxImageCount)
}

function startGenerationTimer(startedAtMs = Date.now()): void {
  stopGenerationTimer()
  const startedAt = Number.isFinite(startedAtMs) ? startedAtMs : Date.now()
  elapsedSeconds.value = Math.max(0, Math.floor((Date.now() - startedAt) / 1000))
  waitingStepIndex.value = 0
  generationTimerId = setInterval(() => {
    elapsedSeconds.value = Math.max(0, Math.floor((Date.now() - startedAt) / 1000))
    waitingStepIndex.value = Math.floor(elapsedSeconds.value / 8) % waitingStepKeys.length
  }, 1000)
}

function stopGenerationTimer(): void {
  if (!generationTimerId) return
  clearInterval(generationTimerId)
  generationTimerId = null
}

function stopTaskPolling(): void {
  if (!taskPollTimerId) return
  clearInterval(taskPollTimerId)
  taskPollTimerId = null
}

function startTaskPolling(taskId: number, messageId: string | null = null, startedAtMs = Date.now()): void {
  activeTaskId.value = taskId
  activePollMessageId = messageId
  generating.value = true
  startGenerationTimer(startedAtMs)
  stopTaskPolling()
  void pollImageTask(taskId)
  taskPollTimerId = setInterval(() => {
    void pollImageTask(taskId)
  }, 2500)
}

async function pollImageTask(taskId: number): Promise<void> {
  try {
    const task = await getImageTask(taskId)
    if (activeTaskId.value !== taskId) return
    upsertTask(task)

    const message = findMessageById(activePollMessageId)
    if (message) {
      message.status = task.status
      message.taskId = task.id
    }

    if (taskIsActive(task)) {
      return
    }

    stopTaskPolling()
    stopGenerationTimer()
    generating.value = false
    activeTaskId.value = null
    activePollMessageId = null

    if (task.status === 'succeeded') {
      const images = (task.images || []).map(storedImageToResult)
      mergeGalleryImages(images)
      if (message) {
        message.content = task.prompt
        message.images = images
      }
      await hydrateImages(images)
      appStore.showSuccess(t('chatImageStudio.generateSuccess', { count: images.length }))
      void scrollToBottom()
      return
    }

    const errorMessage = task.error_message || t('chatImageStudio.generateFailed')
    if (message) {
      message.content = errorMessage
    }
    appStore.showError(errorMessage)
  } catch (error: any) {
    const message = findMessageById(activePollMessageId)
    stopTaskPolling()
    stopGenerationTimer()
    generating.value = false
    activeTaskId.value = null
    activePollMessageId = null
    if (message) {
      message.status = 'failed'
      message.content = error?.message || t('chatImageStudio.generateFailed')
    }
    appStore.showError(error?.message || t('chatImageStudio.generateFailed'))
  }
}

function findMessageById(id: string | null): StudioMessage | null {
  if (!id) return null
  for (const session of sessions.value) {
    const message = session.messages.find((item) => item.id === id)
    if (message) return message
  }
  return null
}

function taskIsActive(task: ImageCreatorTask | null | undefined): boolean {
  return task?.status === 'pending' || task?.status === 'running'
}

function taskIsActiveStatus(status: ImageCreatorTaskStatus | undefined): boolean {
  return status === 'pending' || status === 'running'
}

function imageStatusLabel(status: ImageCreatorTaskStatus | undefined): string {
  if (!status) return t('chatImageStudio.status.pending')
  return t(`chatImageStudio.status.${status}`)
}

function upsertTask(task: ImageCreatorTask): void {
  const index = imageTasks.value.findIndex((item) => item.id === task.id)
  if (index >= 0) {
    imageTasks.value[index] = task
  } else {
    imageTasks.value.unshift(task)
  }
  imageTasks.value = imageTasks.value
    .slice()
    .sort((a, b) => Date.parse(b.created_at) - Date.parse(a.created_at))
}

async function loadImageTasks(): Promise<void> {
  loadingImages.value = true
  try {
    const response = await listImageTasks()
    imageTasks.value = (response.tasks || [])
      .slice()
      .sort((a, b) => Date.parse(b.created_at) - Date.parse(a.created_at))
    const images = (response.images?.length ? response.images : imagesFromTasks(response.tasks || [])).map(storedImageToResult)
    galleryImages.value = images
    await hydrateImages(images)
    await syncImageMessagesFromTasks(imageTasks.value)
    const active = imageTasks.value.find(taskIsActive)
    if (active && !generating.value) {
      const activeMessage = findImageMessageByTaskId(active.id)
      startTaskPolling(active.id, activeMessage?.id ?? null, taskTimerStartMs(active))
    }
  } catch (error: any) {
    appStore.showError(error?.message || t('chatImageStudio.loadTasksFailed'))
  } finally {
    loadingImages.value = false
  }
}

function imagesFromTasks(tasks: ImageCreatorTask[]): ImageCreatorStoredImage[] {
  return tasks.flatMap((task) => task.images || [])
}

function storedImageToResult(image: ImageCreatorStoredImage, index: number): GeneratedImage {
  return {
    id: String(image.id || `${Date.now()}-${index}`),
    url: image.url,
    sourceUrl: image.url,
    revisedPrompt: image.revised_prompt || '',
    outputFormat: image.output_format || outputFormat.value,
    mimeType: image.mime_type || '',
    byteSize: image.byte_size,
  }
}

function mergeGalleryImages(images: GeneratedImage[]): void {
  const byId = new Map<string, GeneratedImage>()
  for (const image of images) byId.set(image.id, image)
  for (const image of galleryImages.value) {
    if (!byId.has(image.id)) byId.set(image.id, image)
  }
  galleryImages.value = Array.from(byId.values())
}

function taskTimerStartMs(task: ImageCreatorTask | null | undefined): number {
  const raw = task?.started_at || task?.created_at
  const parsed = raw ? Date.parse(raw) : NaN
  return Number.isFinite(parsed) ? parsed : Date.now()
}

function imagesFromSessions(): GeneratedImage[] {
  return sessions.value.flatMap((session) =>
    session.messages.flatMap((message) => message.images || [])
  )
}

async function syncImageMessagesFromTasks(tasks: ImageCreatorTask[]): Promise<void> {
  const imagesToHydrate: GeneratedImage[] = []
  for (const task of tasks) {
    const message = findImageMessageByTaskId(task.id)
    if (!message) continue

    message.taskId = task.id
    message.status = task.status

    if (task.status === 'succeeded') {
      const images = (task.images || []).map(storedImageToResult)
      message.content = task.prompt
      message.images = images
      mergeGalleryImages(images)
      imagesToHydrate.push(...images)
    } else if (task.status === 'failed') {
      message.content = task.error_message || t('chatImageStudio.generateFailed')
      message.images = []
    } else if (!message.content) {
      message.content = t('chatImageStudio.generatingHint')
    }

    const session = sessionForMessage(message.id)
    if (session) touchSession(session)
  }

  if (imagesToHydrate.length > 0) {
    await hydrateImages(imagesToHydrate)
  }
}

function findImageMessageByTaskId(taskId: number): StudioMessage | null {
  for (const session of sessions.value) {
    const message = session.messages.find((item) => item.kind === 'image' && item.taskId === taskId)
    if (message) return message
  }
  return null
}

function shouldFetchImageUrl(url: string): boolean {
  const value = url.trim().toLowerCase()
  return value !== '' && !value.startsWith('data:') && !value.startsWith('blob:')
}

function imageSrc(image: GeneratedImage): string {
  const displayUrl = imageDisplayUrls.value[image.id]
  if (displayUrl) return displayUrl
  const sourceUrl = image.sourceUrl || image.url
  return shouldFetchImageUrl(sourceUrl) ? imagePreviewFallbackSrc : sourceUrl
}

function isImageSelected(image: GeneratedImage): boolean {
  return selectedImageIds.value.includes(image.id)
}

function toggleImageSelection(image: GeneratedImage): void {
  if (isImageSelected(image)) {
    selectedImageIds.value = selectedImageIds.value.filter((id) => id !== image.id)
    return
  }
  selectedImageIds.value = [...selectedImageIds.value, image.id]
}

function selectImages(images: GeneratedImage[] | undefined): void {
  if (!images?.length) return
  const ids = new Set(selectedImageIds.value)
  for (const image of images) ids.add(image.id)
  selectedImageIds.value = Array.from(ids)
}

function clearImageSelection(images?: GeneratedImage[]): void {
  if (!images?.length) {
    selectedImageIds.value = []
    return
  }
  const idsToClear = new Set(images.map((image) => image.id))
  selectedImageIds.value = selectedImageIds.value.filter((id) => !idsToClear.has(id))
}

function selectedCountForImages(images: GeneratedImage[] | undefined): number {
  if (!images?.length) return 0
  return images.filter(isImageSelected).length
}

function selectedImagesForImages(images: GeneratedImage[] | undefined): DownloadableImage[] {
  if (!images?.length) return []
  return images
    .map((image, index) => ({ image, index }))
    .filter(({ image }) => isImageSelected(image))
}

async function hydrateImages(images: GeneratedImage[]): Promise<void> {
  await Promise.all(images.map(async (image) => {
    const sourceUrl = image.sourceUrl || image.url
    if (!shouldFetchImageUrl(sourceUrl)) {
      imageDisplayUrls.value = {
        ...imageDisplayUrls.value,
        [image.id]: sourceUrl,
      }
      return
    }
    try {
      const blob = await downloadImageFile(sourceUrl)
      const objectUrl = URL.createObjectURL(blob)
      generatedImageObjectUrls.add(objectUrl)
      const previous = imageDisplayUrls.value[image.id]
      if (previous?.startsWith('blob:') && generatedImageObjectUrls.delete(previous)) {
        URL.revokeObjectURL(previous)
      }
      imageDisplayUrls.value = {
        ...imageDisplayUrls.value,
        [image.id]: objectUrl,
      }
    } catch {
      const next = { ...imageDisplayUrls.value }
      delete next[image.id]
      imageDisplayUrls.value = next
    }
  }))
}

function revokeGeneratedImageObjectUrls(): void {
  for (const url of generatedImageObjectUrls) {
    URL.revokeObjectURL(url)
  }
  generatedImageObjectUrls.clear()
}

async function ensureImageDownloadUrl(image: GeneratedImage): Promise<string> {
  const current = imageDisplayUrls.value[image.id]
  if (current?.startsWith('blob:') || current?.startsWith('data:')) return current
  await hydrateImages([image])
  const hydrated = imageDisplayUrls.value[image.id]
  if (hydrated?.startsWith('blob:') || hydrated?.startsWith('data:')) return hydrated
  return image.sourceUrl || image.url
}

async function downloadImage(image: GeneratedImage, index: number, notify = true): Promise<void> {
  const href = await ensureImageDownloadUrl(image)
  const link = document.createElement('a')
  link.href = href
  link.download = `image-${Date.now()}-${index + 1}.${String(image.outputFormat || outputFormat.value).toLowerCase()}`
  if (!href.startsWith('data:') && !href.startsWith('blob:')) {
    link.target = '_blank'
    link.rel = 'noopener'
  }
  document.body.appendChild(link)
  link.click()
  link.remove()
  if (notify) {
    appStore.showSuccess(t('chatImageStudio.downloadStarted'))
  }
}

async function downloadImages(items: DownloadableImage[]): Promise<void> {
  if (items.length === 0) return
  appStore.showSuccess(t('chatImageStudio.preparingDownloads', { count: items.length }))
  for (const item of items) {
    await downloadImage(item.image, item.index, false)
  }
}

async function downloadSelectedImages(): Promise<void> {
  await downloadImages(selectedGalleryImages.value)
}

function openTaskPreview(task: ImageCreatorTask): void {
  if (!task.images?.length) return
  const images = task.images.map(storedImageToResult)
  openPreview(images[0], images)
}

function openPreview(image: GeneratedImage, images: GeneratedImage[] = [image]): void {
  previewImages.value = images.length > 0 ? images : [image]
  previewImage.value = image
  previewZoom.value = 1
  void hydrateImages([image])
}

function closePreview(): void {
  previewImage.value = null
  previewImages.value = []
  previewZoom.value = 1
}

function downloadPreviewImage(): void {
  if (!previewImage.value) return
  void downloadImage(previewImage.value, 0)
}

function showPreviewAt(index: number): void {
  const image = previewImages.value[index]
  if (!image) return
  previewImage.value = image
  previewZoom.value = 1
  void hydrateImages([image])
}

function showPreviousPreview(): void {
  if (!canPreviewPrevious.value) return
  showPreviewAt(previewIndex.value - 1)
}

function showNextPreview(): void {
  if (!canPreviewNext.value) return
  showPreviewAt(previewIndex.value + 1)
}

function zoomPreview(delta: number): void {
  previewZoom.value = clampPreviewZoom(previewZoom.value + delta)
}

function resetPreviewZoom(): void {
  previewZoom.value = 1
}

function clampPreviewZoom(value: number): number {
  return Math.min(maxPreviewZoom, Math.max(minPreviewZoom, Number(value.toFixed(2))))
}

function onPreviewImageLoad(event: Event): void {
  const imageElement = event.target as HTMLImageElement
  if (!previewImage.value || !imageElement.naturalWidth || !imageElement.naturalHeight) return
  previewImage.value.width = imageElement.naturalWidth
  previewImage.value.height = imageElement.naturalHeight
}

function onPreviewKeydown(event: KeyboardEvent): void {
  if (previewImage.value && event.key === 'ArrowLeft') {
    event.preventDefault()
    showPreviousPreview()
    return
  }
  if (previewImage.value && event.key === 'ArrowRight') {
    event.preventDefault()
    showNextPreview()
    return
  }
  if (event.key === 'Escape') {
    if (previewImage.value) {
      closePreview()
      return
    }
    if (imageParamsOpen.value) {
      imageParamsOpen.value = false
      return
    }
    if (promptMarketOpen.value) {
      closePromptMarket()
      return
    }
    queueOpen.value = false
  }
}

async function openPromptMarket(): Promise<void> {
  promptMarketOpen.value = true
  mode.value = 'image'
  activeTab.value = 'studio'
  if (promptMarketItems.value.length === 0 && !promptMarketLoading.value) {
    await reloadPromptMarket()
    return
  }
  if (promptMarketFavorites.value.length === 0) {
    await loadPromptFavorites()
  }
}

function closePromptMarket(): void {
  promptMarketOpen.value = false
}

async function reloadPromptMarket(): Promise<void> {
  promptMarketAbortController?.abort()
  const controller = new AbortController()
  promptMarketAbortController = controller
  promptMarketLoading.value = true
  promptMarketError.value = ''
  try {
    const [items] = await Promise.all([
      fetchPromptMarketPrompts(controller.signal),
      loadPromptFavorites(controller.signal),
    ])
    promptMarketItems.value = items
  } catch (error: unknown) {
    if (isAbortError(error)) return
    const message = error instanceof Error ? error.message : t('chatImageStudio.promptMarketLoadFailed')
    promptMarketError.value = message
    appStore.showError(promptMarketError.value)
  } finally {
    if (promptMarketAbortController === controller) {
      promptMarketAbortController = null
      promptMarketLoading.value = false
    }
  }
}

async function loadPromptFavorites(signal?: AbortSignal): Promise<void> {
  const response = await fetchPromptFavorites(signal)
  promptMarketFavorites.value = response.items || []
}

function isPromptFavorited(item: BananaPrompt): boolean {
  return promptFavoriteRecords.value.has(promptFavoriteKey(item))
}

async function togglePromptFavorite(item: BananaPrompt): Promise<void> {
  try {
    const favorite = promptFavoriteRecords.value.get(promptFavoriteKey(item))
    if (favorite) {
      const response = await deletePromptFavorite(favorite.id)
      promptMarketFavorites.value = response.items || []
      return
    }
    const response = await createPromptFavorite(item)
    promptMarketFavorites.value = response.items || []
  } catch (error: unknown) {
    const message = error instanceof Error ? error.message : t('chatImageStudio.promptFavoriteFailed')
    appStore.showError(message)
  }
}

async function applyPromptMarketItem(item: BananaPrompt, withReference: boolean): Promise<void> {
  prompt.value = item.prompt
  mode.value = 'image'
  activeTab.value = 'studio'
  closePromptMarket()
  if (withReference && item.referenceImageUrls.length > 0) {
    await attachPromptMarketReference(item.referenceImageUrls[0])
  }
}

async function attachPromptMarketReference(url: string): Promise<void> {
  try {
    const response = await fetch(url)
    if (!response.ok) throw new Error(`reference image ${response.status}`)
    const blob = await response.blob()
    const mimeType = blob.type || 'image/png'
    if (!['image/png', 'image/jpeg', 'image/webp'].includes(mimeType.toLowerCase())) {
      throw new Error('unsupported reference image')
    }
    const ext = referenceImageExtension(mimeType)
    setReferenceImage(new File([blob], `prompt-reference.${ext}`, { type: mimeType }))
  } catch (error: unknown) {
    const message = error instanceof Error ? error.message : t('chatImageStudio.promptReferenceAttachFailed')
    appStore.showWarning(message)
  }
}

function triggerReferenceUpload(): void {
  referenceInputRef.value?.click()
}

function onReferenceImageChange(event: Event): void {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0] ?? null
  if (!file) return
  setReferenceImage(file)
  input.value = ''
}

function onComposerPaste(event: ClipboardEvent): void {
  const file = pastedReferenceImage(event)
  if (!file) return
  event.preventDefault()
  mode.value = 'image'
  setReferenceImage(file)
  appStore.showSuccess(t('chatImageStudio.referenceAttachedFromPaste'))
}

function pastedReferenceImage(event: ClipboardEvent): File | null {
  const clipboard = event.clipboardData
  if (!clipboard) return null

  const pastedFile = Array.from(clipboard.files || []).find(isSupportedReferenceImage)
  if (pastedFile) return pastedFile

  for (const item of Array.from(clipboard.items || [])) {
    if (item.kind !== 'file') continue
    const file = item.getAsFile()
    if (file && isSupportedReferenceImage(file)) return file
  }

  return null
}

function isSupportedReferenceImage(file: File): boolean {
  return ['image/png', 'image/jpeg', 'image/webp'].includes(file.type.toLowerCase())
}

function setReferenceImage(file: File): void {
  clearReferenceImage()
  referenceImage.value = file
  referencePreviewUrl.value = URL.createObjectURL(file)
}

function clearReferenceImage(): void {
  referenceImage.value = null
  if (referencePreviewUrl.value) {
    URL.revokeObjectURL(referencePreviewUrl.value)
    referencePreviewUrl.value = ''
  }
}

async function scrollToBottom(): Promise<void> {
  await nextTick()
  const el = messagesRef.value
  if (!el) return
  if (typeof el.scrollTo === 'function') {
    el.scrollTo({ top: el.scrollHeight })
  } else {
    el.scrollTop = el.scrollHeight
  }
}

async function copyReply(text: string): Promise<void> {
  await copyToClipboard(text, t('chatImageStudio.replyCopied'))
}

function formatDateTime(value: string): string {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return new Intl.DateTimeFormat(undefined, {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}

function formatBytes(value: number | undefined): string {
  if (typeof value !== 'number' || !Number.isFinite(value) || value <= 0) return ''
  const units = ['B', 'KB', 'MB', 'GB']
  let size = value
  let unitIndex = 0
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024
    unitIndex += 1
  }
  const precision = unitIndex === 0 ? 0 : size >= 10 ? 1 : 2
  return `${size.toFixed(precision)} ${units[unitIndex]}`
}

function formatImageDimensions(image: GeneratedImage): string {
  if (!image.width || !image.height) return ''
  return `${image.width} x ${image.height}`
}

function formatImageAspectRatio(image: GeneratedImage): string {
  if (!image.width || !image.height) return ''
  const divisor = greatestCommonDivisor(image.width, image.height)
  return `${Math.round(image.width / divisor)}: ${Math.round(image.height / divisor)}`.replace(': ', ':')
}

function formatImageMegapixels(image: GeneratedImage): string {
  if (!image.width || !image.height) return ''
  const megapixels = (image.width * image.height) / 1_000_000
  return `${megapixels >= 10 ? megapixels.toFixed(1) : megapixels.toFixed(2)}MP`
}

function greatestCommonDivisor(a: number, b: number): number {
  let x = Math.abs(Math.round(a))
  let y = Math.abs(Math.round(b))
  while (y > 0) {
    const next = x % y
    x = y
    y = next
  }
  return x || 1
}

function updateGalleryColumnCount(): void {
  const width = galleryGridRef.value?.clientWidth || window.innerWidth || 0
  const available = Math.max(160, width - 32)
  const idealCardWidth = available < 720 ? 160 : 220
  galleryColumnCount.value = Math.max(1, Math.min(6, Math.floor(available / idealCardWidth)))
}

function formatDuration(seconds: number): string {
  const value = Math.max(0, Math.floor(seconds))
  const minutes = Math.floor(value / 60)
  const restSeconds = value % 60
  return `${String(minutes).padStart(2, '0')}:${String(restSeconds).padStart(2, '0')}`
}
</script>

<style scoped>
.chat-image-studio {
  display: grid;
  position: relative;
  min-height: calc(100vh - 7rem);
  min-height: calc(100dvh - 7rem);
  grid-template-columns: minmax(15rem, 16.5rem) minmax(0, 1fr);
  gap: 0.75rem;
}

.studio-mobile-backdrop {
  display: none;
}

.studio-rail,
.studio-main {
  height: 100%;
  min-height: 0;
  overflow: hidden;
  border: 1px solid rgb(229 231 235);
  border-radius: 0.5rem;
  background: rgb(248 250 252 / 0.94);
}

.studio-rail {
  background: rgb(255 255 255 / 0.92);
}

.dark .studio-rail,
.dark .studio-main {
  border-color: rgb(55 65 81);
  background: rgb(31 41 55 / 0.86);
}

.studio-rail,
.studio-main,
.studio-assistant-body,
.studio-user-bubble,
.studio-composer-shell {
  box-shadow: 0 18px 48px rgb(15 23 42 / 0.08);
}

.studio-rail,
.studio-main {
  display: flex;
  flex-direction: column;
}

.studio-topbar,
.studio-section-head {
  display: flex;
  flex-shrink: 0;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  border-bottom: 1px solid rgb(243 244 246);
  padding: 0.6rem 0.75rem;
}

.studio-topbar {
  position: relative;
  display: grid;
  grid-template-columns: minmax(15rem, 1fr) auto minmax(15rem, 1fr);
}

.dark .studio-topbar,
.dark .studio-section-head {
  border-color: rgb(55 65 81);
}

.studio-empty-icon,
.studio-avatar,
.studio-generating-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgb(236 253 245);
  color: rgb(5 150 105);
}

.dark .studio-empty-icon,
.dark .studio-avatar,
.dark .studio-generating-icon {
  background: rgb(20 83 45 / 0.35);
  color: rgb(110 231 183);
}

.studio-settings {
  display: grid;
  flex-shrink: 0;
  gap: 0.875rem;
  border-bottom: 1px solid rgb(243 244 246);
  padding: 1rem;
}

.studio-rail-mobile-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.dark .studio-settings {
  border-color: rgb(55 65 81);
}

.studio-topbar-left {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 0.75rem;
}

.studio-rail-actions {
  display: grid;
  gap: 0.625rem;
  flex-shrink: 0;
  padding: 0.875rem 1rem 0;
}

.studio-key-control {
  min-width: 0;
}

.studio-session-list {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 0.75rem;
}

.studio-session-item {
  display: flex;
  width: 100%;
  align-items: flex-start;
  gap: 0.5rem;
  border-radius: 0.5rem;
  padding: 0.25rem;
  color: rgb(55 65 81);
  transition: background-color 0.15s ease, color 0.15s ease;
}

.studio-session-select {
  display: flex;
  min-width: 0;
  flex: 1;
  align-items: flex-start;
  gap: 0.75rem;
  padding: 0.5rem;
  text-align: left;
}

.studio-session-delete,
.studio-icon-action,
.studio-queue-button,
.studio-preview-close {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 0.5rem;
  transition: background-color 0.15s ease, color 0.15s ease, opacity 0.15s ease;
}

.studio-session-delete,
.studio-icon-action {
  height: 2.25rem;
  width: 2.25rem;
  flex-shrink: 0;
  color: rgb(100 116 139);
}

.studio-session-delete:hover:not(:disabled),
.studio-icon-action:hover:not(:disabled),
.studio-queue-button:hover {
  background: rgb(241 245 249);
  color: rgb(15 23 42);
}

.studio-icon-action-active {
  background: rgb(219 234 254);
  color: rgb(37 99 235);
}

.studio-session-delete:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.studio-queue-button {
  position: relative;
  min-height: 1.875rem;
  gap: 0.35rem;
  border-radius: 9999px;
  background: rgb(241 245 249);
  padding: 0.25rem 0.35rem 0.25rem 0.6rem;
  font-size: 0.75rem;
  font-weight: 700;
  color: rgb(71 85 105);
  white-space: nowrap;
}

.studio-queue-label {
  display: inline;
}

.studio-queue-button-active {
  background: rgb(219 234 254);
  color: rgb(37 99 235);
}

.studio-queue-count {
  display: inline-flex;
  min-width: 1.25rem;
  height: 1.25rem;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  background: rgb(255 255 255 / 0.9);
  padding-inline: 0.35rem;
  font-size: 0.6875rem;
  line-height: 1;
}

.studio-session-item:hover,
.studio-session-item-active {
  background: rgb(236 253 245);
  color: rgb(4 120 87);
}

.dark .studio-session-item {
  color: rgb(209 213 219);
}

.dark .studio-session-delete,
.dark .studio-icon-action {
  color: rgb(148 163 184);
}

.dark .studio-session-delete:hover:not(:disabled),
.dark .studio-icon-action:hover:not(:disabled),
.dark .studio-queue-button:hover {
  background: rgb(51 65 85 / 0.65);
  color: rgb(226 232 240);
}

.dark .studio-icon-action-active {
  background: rgb(30 64 175 / 0.35);
  color: rgb(147 197 253);
}

.dark .studio-session-item:hover,
.dark .studio-session-item-active {
  background: rgb(20 83 45 / 0.32);
  color: rgb(167 243 208);
}

.dark .studio-queue-button {
  background: rgb(51 65 85 / 0.75);
  color: rgb(203 213 225);
}

.dark .studio-queue-button-active {
  background: rgb(30 64 175 / 0.35);
  color: rgb(147 197 253);
}

.dark .studio-queue-count {
  background: rgb(15 23 42 / 0.75);
}

.studio-tabs,
.studio-mode-switch,
.studio-mode-cluster {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  border: 1px solid rgb(226 232 240);
  border-radius: 9999px;
  background: rgb(248 250 252);
  padding: 0.25rem;
}

.studio-topbar > .studio-tabs {
  justify-self: center;
}

.dark .studio-tabs,
.dark .studio-mode-switch,
.dark .studio-mode-cluster {
  border-color: rgb(55 65 81);
  background: rgb(17 24 39 / 0.55);
}

.studio-mode-switch {
  border: 0;
  background: transparent;
  padding: 0;
}

.dark .studio-mode-switch {
  border-color: transparent;
  background: transparent;
}

.studio-tab,
.studio-mode-button {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  border-radius: 9999px;
  padding: 0.375rem 0.75rem;
  font-size: 0.8125rem;
  font-weight: 600;
  color: rgb(71 85 105);
  white-space: nowrap;
}

.studio-tab-active,
.studio-mode-button-active {
  background: rgb(255 255 255);
  color: rgb(15 23 42);
  box-shadow: 0 6px 16px rgb(15 23 42 / 0.08);
}

.dark .studio-tab,
.dark .studio-mode-button {
  color: rgb(203 213 225);
}

.dark .studio-tab-active,
.dark .studio-mode-button-active {
  background: rgb(55 65 81);
  color: rgb(255 255 255);
}

.studio-status {
  display: flex;
  min-width: 0;
  align-items: center;
  justify-content: flex-end;
  justify-self: end;
  gap: 0.5rem;
}

.studio-status-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  border-radius: 9999px;
  background: rgb(241 245 249);
  padding: 0.25rem 0.55rem;
  font-size: 0.75rem;
  font-weight: 600;
  color: rgb(71 85 105);
  white-space: nowrap;
}

.studio-status-chip-live {
  background: rgb(219 234 254);
  color: rgb(37 99 235);
}

.dark .studio-status-chip {
  background: rgb(51 65 85 / 0.75);
  color: rgb(203 213 225);
}

.dark .studio-status-chip-live {
  background: rgb(30 64 175 / 0.35);
  color: rgb(147 197 253);
}

.studio-messages,
.studio-gallery {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
}

.studio-messages {
  background: rgb(241 245 249 / 0.62);
  padding: 0 1rem 0.5rem;
}

.dark .studio-messages {
  background: rgb(15 23 42 / 0.22);
}

.studio-gallery {
  padding: 0 0 1rem;
}

.studio-empty {
  display: flex;
  min-height: 420px;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  text-align: center;
}

.studio-empty-compact {
  min-height: 360px;
}

.studio-empty-icon {
  margin-bottom: 1rem;
  height: 4rem;
  width: 4rem;
  border-radius: 0.75rem;
}

.studio-message {
  display: flex;
  width: 100%;
  gap: 0.875rem;
}

.studio-message-stack {
  display: flex;
  width: 100%;
  max-width: 70rem;
  flex-direction: column;
  gap: 0.875rem;
  margin: 0 auto;
  padding: 0.875rem 0;
}

.studio-message-user {
  justify-content: flex-end;
}

.studio-message-assistant {
  align-items: flex-start;
  justify-content: flex-start;
}

.studio-avatar {
  height: 2rem;
  width: 2rem;
  flex-shrink: 0;
  border-radius: 0.5rem;
}

.studio-assistant-body {
  min-width: 0;
  width: min(720px, calc(72% - 3rem));
  border: 1px solid rgb(226 232 240);
  border-radius: 0.5rem;
  background: rgb(255 255 255);
  padding: 0.75rem 0.875rem;
}

.studio-message-kind-image .studio-assistant-body {
  width: min(860px, calc(100% - 3rem));
}

.dark .studio-assistant-body {
  border-color: rgb(55 65 81);
  background: rgb(17 24 39 / 0.5);
}

.studio-user-bubble {
  width: fit-content;
  min-width: min(34rem, 58%);
  max-width: min(900px, 84%);
  border-radius: 0.5rem;
  background: rgb(236 253 245);
  padding: 0.625rem 0.875rem;
  color: rgb(30 41 59);
}

.dark .studio-user-bubble {
  background: rgb(20 83 45 / 0.35);
  color: rgb(240 253 244);
}

.studio-message-toolbar,
.studio-image-card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.studio-image-batchbar,
.studio-gallery-batchbar,
.studio-section-actions,
.studio-image-batch-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.studio-section-actions {
  justify-content: flex-end;
  flex-wrap: wrap;
}

.studio-image-batchbar,
.studio-gallery-batchbar {
  flex-wrap: wrap;
  border-radius: 9999px;
  background: rgb(248 250 252);
  padding: 0.35rem 0.5rem 0.35rem 0.75rem;
  font-size: 0.75rem;
  font-weight: 700;
  color: rgb(71 85 105);
}

.studio-image-batchbar {
  justify-content: space-between;
  border-radius: 0.5rem;
  margin-bottom: 0.75rem;
}

.studio-image-batch-actions {
  flex-wrap: wrap;
  justify-content: flex-end;
}

.studio-text-action {
  display: inline-flex;
  min-height: 1.75rem;
  align-items: center;
  gap: 0.25rem;
  border-radius: 9999px;
  padding: 0.2rem 0.55rem;
  color: rgb(37 99 235);
  transition: background-color 0.15s ease, color 0.15s ease, opacity 0.15s ease;
}

.studio-text-action:hover:not(:disabled) {
  background: rgb(219 234 254);
}

.studio-text-action:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

.studio-message-toolbar {
  margin-bottom: 0.5rem;
  font-size: 0.75rem;
  font-weight: 600;
  color: rgb(100 116 139);
}

.studio-message-toolbar-user {
  color: rgb(71 85 105);
}

.studio-message-toolbar-user .studio-message-action {
  height: 1.75rem;
  width: 1.75rem;
  justify-content: center;
  padding: 0;
}

.dark .studio-message-toolbar {
  color: rgb(148 163 184);
}

.dark .studio-image-batchbar,
.dark .studio-gallery-batchbar {
  background: rgb(15 23 42 / 0.45);
  color: rgb(203 213 225);
}

.dark .studio-text-action {
  color: rgb(147 197 253);
}

.dark .studio-text-action:hover:not(:disabled) {
  background: rgb(30 64 175 / 0.35);
}

.studio-message-actions {
  display: inline-flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: flex-end;
  gap: 0.35rem;
}

.studio-message-action {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  border: 1px solid transparent;
  border-radius: 9999px;
  padding: 0.25rem 0.5rem;
  color: rgb(100 116 139);
  transition: background-color 0.15s ease, border-color 0.15s ease, color 0.15s ease, opacity 0.15s ease;
}

.studio-message-action:hover:not(:disabled) {
  border-color: rgb(226 232 240);
  background: rgb(241 245 249);
  color: rgb(30 41 59);
}

.studio-message-action:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.dark .studio-message-action:hover:not(:disabled) {
  border-color: rgb(55 65 81);
  background: rgb(51 65 85 / 0.65);
  color: rgb(226 232 240);
}

.studio-message-content {
  font-size: 0.925rem;
  line-height: 1.75;
  color: rgb(17 24 39);
}

.dark .studio-message-content {
  color: rgb(243 244 246);
}

.studio-typing {
  color: rgb(107 114 128);
}

.studio-image-grid {
  display: grid;
  gap: 0.875rem;
}

.studio-image-grid {
  grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
}

.studio-gallery-grid {
  display: grid;
  align-items: start;
  justify-content: start;
  gap: 0.875rem;
  padding: 1rem;
}

.studio-gallery-column {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 0.875rem;
}

.studio-image-card,
.studio-gallery-card,
.studio-task-row {
  position: relative;
  overflow: hidden;
  border: 1px solid rgb(226 232 240);
  border-radius: 0.5rem;
  background: rgb(255 255 255);
}

.studio-image-card-selected {
  border-color: rgb(37 99 235);
  box-shadow: 0 0 0 3px rgb(59 130 246 / 0.18);
}

.studio-gallery-card {
  width: 100%;
  box-shadow: 0 14px 32px rgb(15 23 42 / 0.08);
}

.dark .studio-image-card,
.dark .studio-gallery-card,
.dark .studio-task-row {
  border-color: rgb(55 65 81);
  background: rgb(17 24 39 / 0.45);
}

.studio-image-preview {
  display: flex;
  width: 100%;
  align-items: center;
  justify-content: center;
  background: rgb(241 245 249);
}

.studio-gallery-preview {
  display: block;
  width: 100%;
  background: rgb(241 245 249);
}

.studio-image-select-toggle {
  position: absolute;
  top: 0.5rem;
  left: 0.5rem;
  z-index: 2;
  display: inline-flex;
  height: 1.6rem;
  width: 1.6rem;
  align-items: center;
  justify-content: center;
  border: 1px solid rgb(255 255 255 / 0.9);
  border-radius: 9999px;
  background: rgb(15 23 42 / 0.58);
  color: rgb(255 255 255);
  box-shadow: 0 8px 18px rgb(15 23 42 / 0.2);
  transition: background-color 0.15s ease, border-color 0.15s ease, transform 0.15s ease;
}

.studio-image-select-toggle:hover {
  transform: translateY(-1px);
  background: rgb(37 99 235);
}

.studio-image-select-toggle-active {
  border-color: rgb(37 99 235);
  background: rgb(37 99 235);
}

.studio-image-select-dot {
  height: 0.45rem;
  width: 0.45rem;
  border-radius: 9999px;
  background: rgb(255 255 255 / 0.85);
}

.studio-image-preview {
  aspect-ratio: 1 / 1;
}

.dark .studio-image-preview,
.dark .studio-gallery-preview {
  background: rgb(15 23 42 / 0.7);
}

.studio-image-preview img {
  height: 100%;
  width: 100%;
  object-fit: contain;
}

.studio-gallery-preview img {
  display: block;
  height: auto;
  width: 100%;
  object-fit: contain;
}

.studio-gallery-download {
  position: absolute;
  right: 0.5rem;
  bottom: 0.5rem;
  z-index: 2;
  display: inline-flex;
  height: 2rem;
  width: 2rem;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  background: rgb(15 23 42 / 0.68);
  color: rgb(255 255 255);
  opacity: 0;
  box-shadow: 0 10px 22px rgb(15 23 42 / 0.24);
  transform: translateY(0.25rem);
  transition: opacity 0.15s ease, transform 0.15s ease, background-color 0.15s ease;
}

.studio-gallery-card:hover .studio-gallery-download,
.studio-gallery-card:focus-within .studio-gallery-download {
  opacity: 1;
  transform: translateY(0);
}

.studio-gallery-download:hover {
  background: rgb(37 99 235);
}

.studio-image-card-footer {
  padding: 0.65rem 0.75rem;
  font-size: 0.75rem;
  font-weight: 600;
  color: rgb(100 116 139);
}

.studio-image-card-footer > span {
  border-radius: 9999px;
  background: rgb(241 245 249);
  padding: 0.15rem 0.45rem;
}

.dark .studio-image-card-footer > span {
  background: rgb(51 65 85 / 0.75);
}

.studio-generating {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  border-radius: 0.5rem;
  background: rgb(248 250 252);
  padding: 0.875rem;
}

.dark .studio-generating {
  background: rgb(15 23 42 / 0.45);
}

.studio-generating-icon {
  height: 2.75rem;
  width: 2.75rem;
  flex-shrink: 0;
  border-radius: 0.5rem;
}

.studio-generating-preview {
  position: relative;
  display: flex;
  aspect-ratio: 1 / 1;
  width: min(9rem, 34vw);
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border-radius: 0.5rem;
  background: linear-gradient(135deg, rgb(226 232 240), rgb(248 250 252));
}

.studio-generating-shine {
  position: absolute;
  inset: -35%;
  background: linear-gradient(90deg, transparent, rgb(255 255 255 / 0.75), transparent);
  transform: rotate(18deg);
  animation: studio-shine 1.6s infinite;
}

.dark .studio-generating-preview {
  background: linear-gradient(135deg, rgb(30 41 59), rgb(15 23 42));
}

.dark .studio-generating-shine {
  background: linear-gradient(90deg, transparent, rgb(148 163 184 / 0.18), transparent);
}

@keyframes studio-shine {
  0% {
    translate: -80% 0;
  }
  100% {
    translate: 80% 0;
  }
}

.studio-task-list {
  display: grid;
  gap: 0.75rem;
  padding: 1rem;
}

.studio-task-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.875rem 1rem;
}

.studio-queue-overlay {
  position: fixed;
  inset: 0;
  z-index: 65;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding: clamp(4rem, 9vh, 6rem) 1rem 1rem;
  background: rgb(15 23 42 / 0.18);
  backdrop-filter: blur(2px);
}

.studio-queue-modal {
  display: flex;
  width: min(100%, 34rem);
  max-height: min(70vh, 34rem);
  flex-direction: column;
  overflow: hidden;
  border: 1px solid rgb(226 232 240);
  border-radius: 0.75rem;
  background: rgb(255 255 255 / 0.96);
  box-shadow: 0 24px 80px rgb(15 23 42 / 0.2);
}

.dark .studio-queue-modal {
  border-color: rgb(55 65 81);
  background: rgb(17 24 39 / 0.96);
}

.studio-queue-modal-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  border-bottom: 1px solid rgb(243 244 246);
  padding: 1rem;
}

.dark .studio-queue-modal-head {
  border-color: rgb(55 65 81);
}

.studio-queue-modal-icon {
  display: inline-flex;
  height: 2rem;
  width: 2rem;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  border-radius: 0.6rem;
  background: rgb(219 234 254);
  color: rgb(37 99 235);
}

.dark .studio-queue-modal-icon {
  background: rgb(30 64 175 / 0.35);
  color: rgb(147 197 253);
}

.studio-queue-modal-body {
  min-height: 0;
  flex: 1;
  overflow-y: auto;
}

.studio-prompt-market-overlay {
  position: fixed;
  inset: 0;
  z-index: 70;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding: clamp(3rem, 7vh, 5rem) 1rem 1rem;
  background: rgb(15 23 42 / 0.2);
  backdrop-filter: blur(2px);
}

.studio-prompt-market-panel {
  display: flex;
  width: min(100%, 58rem);
  max-height: min(82vh, 44rem);
  flex-direction: column;
  overflow: hidden;
  border: 1px solid rgb(226 232 240);
  border-radius: 0.75rem;
  background: rgb(255 255 255 / 0.98);
  box-shadow: 0 24px 90px rgb(15 23 42 / 0.22);
}

.dark .studio-prompt-market-panel {
  border-color: rgb(55 65 81);
  background: rgb(17 24 39 / 0.98);
}

.studio-prompt-market-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  border-bottom: 1px solid rgb(243 244 246);
  padding: 1rem;
}

.studio-prompt-market-header h2 {
  font-size: 1rem;
  font-weight: 800;
  color: rgb(15 23 42);
}

.studio-prompt-market-header p {
  margin-top: 0.25rem;
  font-size: 0.8125rem;
  color: rgb(100 116 139);
}

.dark .studio-prompt-market-header {
  border-color: rgb(55 65 81);
}

.dark .studio-prompt-market-header h2 {
  color: rgb(248 250 252);
}

.dark .studio-prompt-market-header p {
  color: rgb(148 163 184);
}

.studio-prompt-market-filters {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.75rem;
  border-bottom: 1px solid rgb(243 244 246);
  padding: 0.75rem 1rem;
}

.dark .studio-prompt-market-filters {
  border-color: rgb(55 65 81);
}

.studio-prompt-market-search {
  display: inline-flex;
  min-width: min(100%, 18rem);
  flex: 1 1 18rem;
  align-items: center;
  gap: 0.5rem;
  border: 1px solid rgb(209 213 219);
  border-radius: 9999px;
  background: rgb(255 255 255);
  padding: 0 0.75rem;
  color: rgb(100 116 139);
}

.studio-prompt-market-search input {
  min-width: 0;
  width: 100%;
  border: 0;
  background: transparent;
  padding: 0.55rem 0;
  font-size: 0.875rem;
  color: rgb(15 23 42);
  outline: none;
}

.dark .studio-prompt-market-search {
  border-color: rgb(75 85 99);
  background: rgb(15 23 42 / 0.7);
  color: rgb(148 163 184);
}

.dark .studio-prompt-market-search input {
  color: rgb(248 250 252);
}

.studio-prompt-market-source {
  min-width: 13rem;
}

.studio-prompt-market-error {
  margin: 0.75rem 1rem 0;
  border: 1px solid rgb(254 202 202);
  border-radius: 0.5rem;
  background: rgb(254 242 242);
  padding: 0.65rem 0.75rem;
  font-size: 0.8125rem;
  color: rgb(185 28 28);
}

.dark .studio-prompt-market-error {
  border-color: rgb(127 29 29);
  background: rgb(127 29 29 / 0.25);
  color: rgb(254 202 202);
}

.studio-prompt-market-empty {
  display: flex;
  min-height: 16rem;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  font-size: 0.875rem;
  color: rgb(100 116 139);
}

.dark .studio-prompt-market-empty {
  color: rgb(148 163 184);
}

.studio-prompt-market-list {
  display: grid;
  min-height: 0;
  grid-template-columns: repeat(auto-fill, minmax(17rem, 1fr));
  gap: 0.75rem;
  overflow-y: auto;
  padding: 1rem;
}

.studio-prompt-market-card {
  display: grid;
  grid-template-columns: 5.75rem minmax(0, 1fr);
  gap: 0.75rem;
  overflow: hidden;
  border: 1px solid rgb(226 232 240);
  border-radius: 0.5rem;
  background: rgb(248 250 252);
  padding: 0.75rem;
}

.dark .studio-prompt-market-card {
  border-color: rgb(55 65 81);
  background: rgb(15 23 42 / 0.55);
}

.studio-prompt-market-card img {
  aspect-ratio: 1 / 1;
  width: 5.75rem;
  border-radius: 0.45rem;
  object-fit: cover;
  background: rgb(226 232 240);
}

.studio-prompt-market-card-body {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 0.45rem;
}

.studio-prompt-market-card-title-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.5rem;
}

.studio-prompt-market-card h3 {
  min-width: 0;
  font-size: 0.875rem;
  font-weight: 800;
  line-height: 1.35;
  color: rgb(15 23 42);
}

.dark .studio-prompt-market-card h3 {
  color: rgb(248 250 252);
}

.studio-prompt-market-favorite {
  display: inline-flex;
  height: 1.9rem;
  width: 1.9rem;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  border: 1px solid rgb(209 213 219);
  border-radius: 9999px;
  color: rgb(100 116 139);
}

.studio-prompt-market-favorite-active,
.studio-prompt-market-favorite:hover {
  border-color: rgb(37 99 235);
  background: rgb(219 234 254);
  color: rgb(37 99 235);
}

.dark .studio-prompt-market-favorite {
  border-color: rgb(75 85 99);
  color: rgb(148 163 184);
}

.dark .studio-prompt-market-favorite-active,
.dark .studio-prompt-market-favorite:hover {
  border-color: rgb(59 130 246);
  background: rgb(30 64 175 / 0.35);
  color: rgb(147 197 253);
}

.studio-prompt-market-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
}

.studio-prompt-market-meta span {
  border-radius: 9999px;
  background: rgb(226 232 240);
  padding: 0.1rem 0.4rem;
  font-size: 0.6875rem;
  font-weight: 800;
  color: rgb(71 85 105);
}

.dark .studio-prompt-market-meta span {
  background: rgb(51 65 85 / 0.75);
  color: rgb(203 213 225);
}

.studio-prompt-market-card p {
  display: -webkit-box;
  overflow: hidden;
  color: rgb(71 85 105);
  font-size: 0.8125rem;
  line-height: 1.55;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
}

.dark .studio-prompt-market-card p {
  color: rgb(203 213 225);
}

.studio-prompt-market-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.45rem;
  padding-top: 0.1rem;
}

.studio-queue-empty {
  display: flex;
  min-height: 18rem;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  text-align: center;
}

.studio-composer {
  flex-shrink: 0;
  padding: 0.75rem 1rem 1rem;
}

.studio-composer-shell {
  margin: 0 auto;
  width: min(100%, 58rem);
  border: 1px solid rgb(209 213 219);
  border-radius: 0.5rem;
  background: rgb(255 255 255);
  padding: 0.75rem 0.875rem;
  position: relative;
}

.dark .studio-composer-shell {
  border-color: rgb(75 85 99);
  background: rgb(17 24 39 / 0.65);
}

.studio-input {
  width: 100%;
  min-height: 66px;
  max-height: 180px;
  resize: vertical;
  border: 0;
  background: transparent;
  font-size: 0.925rem;
  line-height: 1.6;
  color: rgb(17 24 39);
  outline: none;
}

.dark .studio-input {
  color: rgb(243 244 246);
}

.studio-composer-toolbar {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding-top: 0.5rem;
}

.studio-submit-group {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 0.5rem;
}

.studio-chat-model-control,
.studio-image-controls {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: nowrap;
}

.studio-chat-model-control {
  flex: 0 1 auto;
  max-width: 100%;
}

.studio-image-controls {
  flex-wrap: nowrap;
  width: auto;
}

.studio-control-label {
  flex-shrink: 0;
  font-size: 0.75rem;
  font-weight: 700;
  color: rgb(100 116 139);
}

.dark .studio-control-label {
  color: rgb(148 163 184);
}

.studio-control-field {
  display: grid;
  min-width: 6.25rem;
  gap: 0.25rem;
  flex: 0 0 auto;
}

.studio-control-field > span {
  font-size: 0.6875rem;
  font-weight: 700;
  line-height: 1;
  color: rgb(100 116 139);
}

.studio-control-field-wide {
  min-width: 9rem;
}

.studio-control-field-count {
  min-width: 3.75rem;
}

.studio-image-params-popover {
  position: absolute;
  right: 4.5rem;
  bottom: calc(100% - 0.25rem);
  z-index: 12;
  width: min(100% - 2rem, 32rem);
  border: 1px solid rgb(209 213 219);
  border-radius: 0.75rem;
  background: rgb(248 250 252 / 0.98);
  padding: 0.75rem;
  box-shadow: 0 20px 70px rgb(15 23 42 / 0.18);
}

.dark .studio-image-params-popover {
  border-color: rgb(55 65 81);
  background: rgb(17 24 39 / 0.98);
}

.studio-image-params-callout {
  border: 1px solid rgb(191 219 254);
  border-radius: 0.6rem;
  background: rgb(239 246 255);
  padding: 0.55rem 0.7rem;
  font-size: 0.75rem;
  line-height: 1.6;
  color: rgb(51 65 85);
}

.studio-image-params-callout strong {
  font-size: 0.8125rem;
  color: rgb(30 41 59);
}

.dark .studio-image-params-callout {
  border-color: rgb(30 64 175 / 0.7);
  background: rgb(30 64 175 / 0.22);
  color: rgb(203 213 225);
}

.dark .studio-image-params-callout strong {
  color: rgb(239 246 255);
}

.studio-image-params-grid {
  margin-top: 0.65rem;
  display: flex;
  flex-wrap: wrap;
  align-items: end;
  gap: 0.5rem;
}

.studio-image-params-note {
  margin-top: 0.65rem;
  font-size: 0.75rem;
  line-height: 1.6;
  color: rgb(100 116 139);
}

.dark .studio-image-params-note {
  color: rgb(148 163 184);
}

.dark .studio-control-field > span {
  color: rgb(148 163 184);
}

.studio-submit-group {
  margin-left: auto;
  justify-content: flex-end;
}

.studio-circle-action,
.studio-send-button {
  display: inline-flex;
  height: 2.5rem;
  width: 2.5rem;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  transition: background-color 0.15s ease, color 0.15s ease, opacity 0.15s ease, transform 0.15s ease;
}

.studio-circle-action {
  border: 1px solid rgb(209 213 219);
  background: rgb(255 255 255);
  color: rgb(71 85 105);
}

.studio-circle-action:hover:not(:disabled),
.studio-circle-action-active {
  border-color: rgb(37 99 235);
  background: rgb(219 234 254);
  color: rgb(37 99 235);
}

.studio-send-button {
  background: rgb(37 99 235);
  color: rgb(255 255 255);
  box-shadow: 0 10px 24px rgb(37 99 235 / 0.3);
}

.studio-send-button:hover:not(:disabled) {
  background: rgb(29 78 216);
  transform: translateY(-1px);
}

.studio-send-button:disabled,
.studio-circle-action:disabled {
  cursor: not-allowed;
  opacity: 0.45;
  transform: none;
  box-shadow: none;
}

.dark .studio-circle-action {
  border-color: rgb(75 85 99);
  background: rgb(17 24 39 / 0.65);
  color: rgb(203 213 225);
}

.dark .studio-circle-action:hover:not(:disabled),
.dark .studio-circle-action-active {
  border-color: rgb(59 130 246);
  background: rgb(30 64 175 / 0.35);
  color: rgb(147 197 253);
}

.studio-chat-model-select {
  flex: 1 1 auto;
  min-width: 9rem;
  max-width: 13rem;
}

.studio-inline-model-select {
  min-width: 9.5rem;
  max-width: 13rem;
}

.studio-mode-model-control {
  min-width: 0;
}

.studio-mode-model-select {
  min-width: 8.5rem;
  max-width: 12rem;
}

.studio-mode-model-select :deep(.select-trigger) {
  min-height: 2rem;
  align-items: center;
  border: 0;
  border-left: 1px solid rgb(226 232 240);
  border-radius: 0;
  background: transparent;
  padding: 0.25rem 0.45rem 0.25rem 0.7rem;
  box-shadow: none;
}

.studio-mode-model-select :deep(.select-value) {
  display: inline-flex;
  min-width: 0;
  align-items: center;
  font-weight: 700;
  font-size: 0.8125rem;
  color: rgb(51 65 85);
}

.studio-model-selected {
  display: inline-flex;
  min-width: 0;
  align-items: center;
  gap: 0.35rem;
}

.studio-model-selected-label {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.studio-mode-model-select :deep(.select-icon) {
  color: rgb(100 116 139);
}

.dark .studio-mode-model-select :deep(.select-trigger) {
  border-left-color: rgb(55 65 81);
}

.dark .studio-mode-model-select :deep(.select-value) {
  color: rgb(226 232 240);
}

.dark .studio-mode-model-select :deep(.select-icon) {
  color: rgb(148 163 184);
}

.studio-tool-button {
  display: inline-flex;
  height: 2.25rem;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  gap: 0.35rem;
  border: 1px solid rgb(209 213 219);
  border-radius: 9999px;
  background: rgb(255 255 255);
  padding: 0 0.7rem;
  font-size: 0.8125rem;
  font-weight: 700;
  color: rgb(71 85 105);
  transition: border-color 0.15s ease, background-color 0.15s ease, color 0.15s ease;
  white-space: nowrap;
}

.studio-tool-button-icon {
  width: 2.25rem;
  padding: 0;
}

.studio-params-button {
  border-color: transparent;
  background: transparent;
  color: rgb(71 85 105);
}

.studio-params-button:hover:not(:disabled) {
  background: rgb(255 255 255 / 0.8);
  color: rgb(37 99 235);
}

.studio-params-button.studio-tool-button-active {
  border-color: transparent;
  background: transparent;
  color: rgb(37 99 235);
}

.studio-tool-button:hover,
.studio-tool-button-active {
  border-color: rgb(37 99 235);
  background: rgb(219 234 254);
  color: rgb(37 99 235);
}

.dark .studio-tool-button {
  border-color: rgb(75 85 99);
  background: rgb(17 24 39 / 0.65);
  color: rgb(203 213 225);
}

.dark .studio-params-button {
  border-color: transparent;
  background: transparent;
  color: rgb(203 213 225);
}

.dark .studio-params-button:hover:not(:disabled),
.dark .studio-params-button.studio-tool-button-active {
  border-color: transparent;
  background: transparent;
  color: rgb(147 197 253);
}

.dark .studio-tool-button:hover,
.dark .studio-tool-button-active {
  border-color: rgb(59 130 246);
  background: rgb(30 64 175 / 0.35);
  color: rgb(147 197 253);
}

.studio-control-select {
  min-width: 9rem;
}

.studio-control-small {
  min-width: 6.5rem;
}

.studio-count-select {
  min-width: 4.5rem;
}

.studio-count-select :deep(.select-trigger) {
  height: 2.25rem;
  min-height: 2.25rem;
  border-radius: 0.5rem;
  padding: 0 0.55rem 0 0.75rem;
}

.studio-count-select :deep(.select-value) {
  text-align: center;
}

.studio-count-select :deep(.select-icon) {
  margin-left: 0.15rem;
}

.studio-reference-bubble {
  position: absolute;
  left: 0.875rem;
  bottom: calc(100% + 0.65rem);
  z-index: 14;
  display: flex;
  width: 4.75rem;
  height: 4.75rem;
  align-items: center;
  justify-content: center;
  border: 1px solid rgb(226 232 240);
  border-radius: 1.25rem;
  background: rgb(255 255 255);
  box-shadow: 0 16px 42px rgb(15 23 42 / 0.16);
}

.dark .studio-reference-bubble {
  border-color: rgb(55 65 81);
  background: rgb(31 41 55);
}

.studio-reference-bubble img {
  width: 100%;
  height: 100%;
  border-radius: inherit;
  object-fit: cover;
}

.studio-reference-remove {
  position: absolute;
  top: -0.45rem;
  right: -0.45rem;
  display: inline-flex;
  width: 1.55rem;
  height: 1.55rem;
  align-items: center;
  justify-content: center;
  border: 1px solid rgb(203 213 225);
  border-radius: 9999px;
  background: rgb(255 255 255 / 0.96);
  color: rgb(100 116 139);
  box-shadow: 0 8px 18px rgb(15 23 42 / 0.18);
  transition:
    background 0.15s ease,
    color 0.15s ease,
    transform 0.15s ease;
}

.studio-reference-remove:hover {
  background: rgb(248 250 252);
  color: rgb(15 23 42);
  transform: translateY(-1px);
}

.dark .studio-reference-remove {
  border-color: rgb(75 85 99);
  background: rgb(17 24 39 / 0.96);
  color: rgb(203 213 225);
}

.studio-preview-overlay {
  position: fixed;
  inset: 0;
  z-index: 9999;
  display: flex;
  flex-direction: column;
  background: rgb(17 17 17 / 0.92);
  padding: 0;
  backdrop-filter: blur(4px);
}

.studio-preview-toolbar {
  position: sticky;
  top: 0;
  z-index: 2;
  display: flex;
  min-height: 3.5rem;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  padding: 0.75rem 4.25rem 0.75rem 1rem;
  background: rgb(15 15 15 / 0.82);
}

.studio-preview-toolbar-center {
  display: flex;
  min-width: 0;
  max-width: 100%;
  align-items: center;
  justify-content: center;
  gap: 0.625rem;
}

.studio-preview-pill {
  display: inline-flex;
  min-height: 2rem;
  align-items: center;
  border-radius: 9999px;
  background: rgb(0 0 0 / 0.72);
  padding: 0.35rem 0.75rem;
  font-size: 0.8125rem;
  font-weight: 700;
  color: rgb(255 255 255 / 0.92);
  box-shadow: 0 12px 28px rgb(0 0 0 / 0.35);
}

.studio-preview-meta {
  max-width: min(52vw, 34rem);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.studio-preview-counter {
  min-width: 4.4rem;
  justify-content: center;
}

.studio-preview-actions {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  border-radius: 9999px;
  background: rgb(0 0 0 / 0.62);
  padding: 0.25rem;
  box-shadow: 0 12px 28px rgb(0 0 0 / 0.35);
}

.studio-preview-tool {
  display: inline-flex;
  height: 2rem;
  width: 2rem;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  color: rgb(255 255 255 / 0.82);
  transition: background-color 0.15s ease, color 0.15s ease;
}

.studio-preview-tool:hover:not(:disabled) {
  background: rgb(255 255 255 / 0.12);
  color: rgb(255 255 255);
}

.studio-preview-tool:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

.studio-preview-close {
  position: absolute;
  top: 0.75rem;
  right: 1rem;
  background: rgb(0 0 0 / 0.45);
}

.studio-preview-nav {
  position: absolute;
  top: 50%;
  z-index: 3;
  display: inline-flex;
  height: 3.35rem;
  width: 3.35rem;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  background: rgb(0 0 0 / 0.62);
  color: rgb(255 255 255 / 0.9);
  box-shadow: 0 18px 46px rgb(0 0 0 / 0.42);
  transform: translateY(-50%);
  transition: background-color 0.15s ease, color 0.15s ease, opacity 0.15s ease, transform 0.15s ease;
}

.studio-preview-nav:hover:not(:disabled) {
  background: rgb(0 0 0 / 0.78);
  color: rgb(255 255 255);
  transform: translateY(-50%) scale(1.04);
}

.studio-preview-nav:disabled {
  cursor: not-allowed;
  opacity: 0.28;
}

.studio-preview-nav-prev {
  left: clamp(2rem, 4vw, 4.5rem);
}

.studio-preview-nav-next {
  right: clamp(2rem, 4vw, 4.5rem);
}

.studio-preview-zoom {
  display: inline-flex;
  min-width: 3.5rem;
  align-items: center;
  justify-content: center;
  padding: 0 0.4rem;
  font-size: 0.8125rem;
  font-weight: 700;
  color: rgb(255 255 255 / 0.92);
}

.studio-preview-body {
  display: flex;
  min-height: 0;
  flex: 1;
  align-items: center;
  justify-content: center;
  overflow: auto;
  padding: 1.25rem;
}

.studio-preview-stage {
  display: flex;
  min-height: min(18rem, 100%);
  width: 100%;
  align-items: center;
  justify-content: center;
}

.studio-preview-stage img {
  display: block;
  max-width: calc(100vw - 2.5rem);
  max-height: calc(100dvh - 6rem);
  border-radius: 0.5rem;
  object-fit: contain;
  box-shadow: 0 24px 80px rgb(0 0 0 / 0.35);
  transform-origin: center center;
  transition: transform 0.15s ease;
}

@media (max-width: 1180px) {
  .studio-topbar {
    grid-template-columns: minmax(12rem, 1fr) auto minmax(12rem, 1fr);
  }

  .studio-composer-toolbar {
    gap: 0.5rem;
  }

  .studio-mode-switch {
    min-width: 0;
  }

  .studio-submit-group {
    justify-content: flex-end;
  }

  .studio-control-field {
    flex-basis: 7.5rem;
  }

  .studio-control-field-wide {
    flex-basis: 9rem;
  }

  .studio-control-field-count {
    flex-basis: 4rem;
  }

}

@media (max-width: 1023px) {
  .chat-image-studio {
    grid-template-columns: 1fr;
  }

  .studio-mobile-backdrop {
    position: fixed;
    inset: 0;
    z-index: 58;
    display: block;
    background: rgb(15 23 42 / 0.42);
    backdrop-filter: blur(2px);
  }

  .studio-rail {
    position: fixed;
    inset: 0 auto 0 0;
    z-index: 59;
    width: min(88vw, 360px);
    max-width: 100%;
    transform: translateX(-105%);
    transition: transform 0.2s ease;
  }

  .studio-rail-mobile-open {
    transform: translateX(0);
  }
}

@media (max-width: 767px) {
  .chat-image-studio {
    min-height: 0;
  }

  .studio-topbar {
    display: flex;
    flex-wrap: nowrap;
    align-items: center;
    border-bottom: 0;
    padding-bottom: 0;
  }

  .studio-topbar-left {
    flex: 0 0 auto;
    width: auto;
  }

  .studio-tabs {
    flex: 1 1 auto;
    min-width: 0;
    justify-content: center;
  }

  .studio-tab {
    flex: 1;
    justify-content: center;
    padding-inline: 0.5rem;
  }

  .studio-status {
    flex: 0 0 auto;
    margin-left: auto;
    width: auto;
    max-width: none;
    overflow: visible;
    align-self: flex-end;
  }

  .studio-queue-button {
    min-height: 2.25rem;
    width: 2.25rem;
    justify-content: center;
    padding: 0;
  }

  .studio-queue-label {
    display: none;
  }

  .studio-queue-count {
    position: absolute;
    top: -0.25rem;
    right: -0.25rem;
    min-width: 1.1rem;
    height: 1.1rem;
    border: 2px solid rgb(255 255 255);
    padding-inline: 0.25rem;
    background: rgb(37 99 235);
    color: rgb(255 255 255);
    font-size: 0.625rem;
  }

  .dark .studio-queue-count {
    border-color: rgb(17 24 39);
    background: rgb(59 130 246);
    color: rgb(255 255 255);
  }

  .studio-messages {
    padding-inline: 0.75rem;
  }

  .studio-message {
    gap: 0.625rem;
  }

  .studio-avatar {
    height: 1.75rem;
    width: 1.75rem;
  }

  .studio-assistant-body {
    width: calc(100% - 2.375rem);
    padding: 0.75rem;
  }

  .studio-user-bubble {
    min-width: 0;
    max-width: 92%;
  }

  .studio-gallery-grid {
    gap: 0.65rem;
  }

  .studio-image-grid {
    grid-template-columns: 1fr;
  }

  .studio-task-row {
    align-items: flex-start;
    flex-direction: column;
  }

  .studio-queue-overlay {
    align-items: stretch;
    padding: 0.75rem;
  }

  .studio-queue-modal {
    width: 100%;
    max-height: calc(100dvh - 1.5rem);
  }

  .studio-queue-modal-head {
    align-items: flex-start;
    flex-direction: column;
  }

  .studio-image-params-popover {
    left: 0.75rem;
    right: 0.75rem;
    width: auto;
  }

  .studio-preview-toolbar {
    flex-wrap: wrap;
    justify-content: center;
    min-height: auto;
    padding: 0.75rem 3.5rem 0.75rem 0.75rem;
  }

  .studio-preview-toolbar-center {
    flex-wrap: wrap;
    justify-content: center;
  }

  .studio-preview-meta {
    max-width: min(100%, calc(100vw - 5rem));
  }

  .studio-preview-body {
    padding: 1rem;
  }

  .studio-preview-nav {
    height: 2.75rem;
    width: 2.75rem;
  }

  .studio-preview-nav-prev {
    left: 1rem;
  }

  .studio-preview-nav-next {
    right: 1rem;
  }

  .studio-preview-stage img {
    max-width: calc(100vw - 2rem);
    max-height: calc(100dvh - 8rem);
  }

  .studio-mode-cluster {
    width: 100%;
  }

  .studio-mode-switch {
    flex: 1 1 auto;
  }

  .studio-mode-button {
    flex: 1;
    justify-content: center;
  }

  .studio-composer-toolbar {
    flex-wrap: wrap;
  }

  .studio-chat-model-control,
  .studio-chat-model-select {
    max-width: none;
  }

  .studio-image-batchbar {
    align-items: flex-start;
    flex-direction: column;
  }

  .studio-generating {
    align-items: flex-start;
    flex-direction: column;
  }

  .studio-generating-preview {
    width: 100%;
  }
}

@media (max-width: 480px) {
  .studio-topbar {
    gap: 0.45rem;
    padding: 0.5rem 0.5rem 0;
  }

  .studio-tab {
    padding-inline: 0.35rem;
  }

  .studio-status {
    gap: 0.35rem;
  }

  .studio-status-chip {
    padding-inline: 0.45rem;
  }

  .studio-composer {
    padding: 0.45rem 0.45rem 0.6rem;
  }

  .studio-composer-shell {
    border-radius: 0.75rem;
    padding: 0.55rem;
    box-shadow: 0 16px 36px rgb(15 23 42 / 0.14);
  }

  .studio-input {
    min-height: 4.5rem;
    max-height: 8.5rem;
    font-size: 0.9rem;
    line-height: 1.55;
  }

  .studio-composer-toolbar {
    display: flex;
    flex-wrap: nowrap;
    gap: 0.35rem;
    align-items: center;
    padding-top: 0.55rem;
  }

  .studio-mode-cluster {
    flex: 1 1 0;
    justify-content: center;
    width: auto;
    min-width: 0;
    min-height: 2.25rem;
    border-radius: 9999px;
    background: rgb(248 250 252);
    padding: 0.16rem 0.28rem;
    gap: 0.12rem;
  }

  .studio-mode-switch {
    flex: 0 0 auto;
    min-width: 0;
    gap: 0.08rem;
  }

  .studio-mode-button {
    flex: 0 0 auto;
    width: 2.2rem;
    min-height: 2rem;
    justify-content: center;
    padding-inline: 0;
    font-size: 0.8125rem;
  }

  .studio-mode-label,
  .studio-model-selected-label,
  .studio-params-label {
    display: none;
  }

  .studio-submit-group {
    flex: 0 0 auto;
    justify-content: flex-end;
    gap: 0.3rem;
  }

  .studio-chat-model-control,
  .studio-image-controls {
    flex: 1 1 0;
    width: auto;
    min-width: 0;
    gap: 0.12rem;
  }

  .studio-chat-model-select,
  .studio-inline-model-select {
    min-width: 2.2rem;
    max-width: none;
    flex: 1 1 0;
  }

  .studio-mode-model-select :deep(.select-trigger) {
    width: 100%;
    height: 2rem;
    min-height: 2rem;
    align-items: center;
    justify-content: center;
    gap: 0;
    border-radius: 9999px;
    padding: 0;
  }

  .studio-mode-model-select :deep(.select-value) {
    display: inline-flex;
    flex: 0 0 100%;
    align-items: center;
    justify-content: center;
    text-align: center;
  }

  .studio-mode-model-select :deep(.select-icon) {
    display: none;
  }

  .studio-model-selected {
    gap: 0;
  }

  .studio-params-button {
    width: 2.1rem;
    padding-inline: 0;
  }

  .studio-tool-button {
    height: 2rem;
    padding-inline: 0.65rem;
  }

  .studio-tool-button-icon,
  .studio-circle-action,
  .studio-send-button {
    width: 2.25rem;
    height: 2.25rem;
  }

  .studio-reference-bubble {
    left: 0.65rem;
    bottom: calc(100% + 0.45rem);
    width: 4rem;
    height: 4rem;
    border-radius: 1rem;
  }
}
</style>
