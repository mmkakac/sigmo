<script setup lang="ts">
import { useI18n } from 'vue-i18n'

import SettingsChannelPicker from '@/components/settings/SettingsChannelPicker.vue'
import SettingsHeader from '@/components/settings/SettingsHeader.vue'
import SettingsNotificationSection from '@/components/settings/SettingsNotificationSection.vue'
import SettingsSaveButton from '@/components/settings/SettingsSaveButton.vue'
import { Label } from '@/components/ui/label'
import { Spinner } from '@/components/ui/spinner'
import {
  TagsInput,
  TagsInputInput,
  TagsInputItem,
  TagsInputItemDelete,
  TagsInputItemText,
} from '@/components/ui/tags-input'
import { useSettingsContext } from '@/composables/useSettingsContext'

const { t } = useI18n()
const {
  channels,
  channelSchemas,
  expandedChannels,
  isReady,
  isSaving,
  isSavingNotifications,
  savingNotificationChannels,
  saveChannel,
  setChannelValue,
  values,
  configuredChannelSchemas,
  saveNotifications,
  toggleChannelDetails,
} = useSettingsContext()

const tagsInputDelimiter = /[,\n\r\t]+/

const cleanKeywords = (keywords: unknown[]) => {
  return keywords.map((line) => String(line).trim()).filter((line) => line.length > 0)
}

</script>

<template>
  <div class="space-y-4">
    <SettingsHeader
      :title="t('settings.notificationTitle')"
      :description="t('settings.notificationDescription')"
    />

    <div
      v-if="!isReady"
      class="flex items-center justify-center py-24"
    >
      <Spinner class="size-6 text-muted-foreground" />
      <span class="sr-only">{{ t('settings.loading') }}</span>
    </div>

    <section
      v-if="isReady && values"
      class="space-y-3 rounded-xl bg-card p-4"
    >
      <h2 class="text-sm font-semibold">{{ t('settings.notificationRouting') }}</h2>
      <p class="text-sm text-muted-foreground">
        {{ t('settings.notificationRoutingDescription') }}
      </p>
      <SettingsChannelPicker
        id="settings-notifications-channels"
        v-model="values.notifications.channels"
        :channels="configuredChannelSchemas"
        :disabled="isSaving"
      />
      <div class="space-y-2 pt-4">
        <Label for="settings-notifications-keywords">
          {{ t('settings.notificationKeywordLabel') }}
        </Label>
        <p class="text-sm text-muted-foreground">
          {{ t('settings.notificationKeywordDescription') }}
        </p>
        <TagsInput
          id="settings-notifications-keywords"
          :model-value="values.notifications.smsFilterKeywords ?? []"
          :disabled="isSaving"
          :delimiter="tagsInputDelimiter"
          add-on-blur
          add-on-paste
          add-on-tab
          @update:model-value="
            values.notifications.smsFilterKeywords = cleanKeywords($event as unknown[])
          "
        >
          <TagsInputItem
            v-for="keyword in values.notifications.smsFilterKeywords ?? []"
            :key="keyword"
            :value="keyword"
          >
            <TagsInputItemText />
            <TagsInputItemDelete />
          </TagsInputItem>
          <TagsInputInput
            :placeholder="t('settings.notificationKeywordPlaceholder')"
            class="min-w-24"
          />
        </TagsInput>
      </div>
      <div class="space-y-2 pt-4">
        <Label for="settings-notifications-numbers">
          {{ t('settings.notificationNumberLabel') }}
        </Label>
        <p class="text-sm text-muted-foreground">
          {{ t('settings.notificationNumberDescription') }}
        </p>
        <TagsInput
          id="settings-notifications-numbers"
          :model-value="values.notifications.smsFilterNumbers ?? []"
          :disabled="isSaving"
          :delimiter="tagsInputDelimiter"
          add-on-blur
          add-on-paste
          add-on-tab
          @update:model-value="
            values.notifications.smsFilterNumbers = cleanKeywords($event as unknown[])
          "
        >
          <TagsInputItem
            v-for="number in values.notifications.smsFilterNumbers ?? []"
            :key="number"
            :value="number"
          >
            <TagsInputItemText />
            <TagsInputItemDelete />
          </TagsInputItem>
          <TagsInputInput
            :placeholder="t('settings.notificationNumberPlaceholder')"
            class="min-w-24"
          />
        </TagsInput>
      </div>
      <div class="flex justify-end border-t pt-4">
        <SettingsSaveButton
          id="settings-notifications-save"
          class="w-full sm:w-auto"
          :disabled="isSaving"
          :saving="isSavingNotifications"
          @save="saveNotifications"
        />
      </div>
    </section>
    <SettingsNotificationSection
      v-if="isReady"
      :channels="channels"
      :disabled="isSaving"
      :expanded-channels="expandedChannels"
      :saving-channels="savingNotificationChannels"
      :schemas="channelSchemas"
      @toggle-details="toggleChannelDetails"
      @update-field="setChannelValue"
      @save-channel="saveChannel"
    />
  </div>
</template>
