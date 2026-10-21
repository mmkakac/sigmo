<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Spinner } from '@/components/ui/spinner'
import { Switch } from '@/components/ui/switch'

const props = defineProps<{
  isLoading: boolean
  isUpdating: boolean
  isValid: boolean
}>()
const emit = defineEmits<{
  (event: 'update'): void
}>()
const alias = defineModel<string>('alias', { required: true })
const mss = defineModel<string>('mss', { required: true })
const autoUnlock = defineModel<boolean>('autoUnlock', { default: false })
const simPin = defineModel<string>('simPin', { default: '' })
const simPinSaved = defineModel<boolean>('simPinSaved', { default: false })

const { t } = useI18n()

const isInputDisabled = computed(() => props.isLoading || props.isUpdating)
const isActionDisabled = computed(() => !props.isValid || props.isUpdating)
</script>

<template>
  <section class="space-y-4 rounded-xl bg-card p-4 shadow-sm">
    <div class="flex items-center justify-between gap-4">
      <h2 class="text-base font-semibold text-foreground">
        {{ t('modemDetail.settings.deviceTitle') }}
      </h2>
    </div>

    <div class="space-y-2">
      <Label for="modem-alias">{{ t('modemDetail.settings.aliasLabel') }}</Label>
      <Input
        id="modem-alias"
        v-model="alias"
        :disabled="isInputDisabled"
        :placeholder="t('modemDetail.settings.aliasPlaceholder')"
      />
      <p class="text-xs text-muted-foreground">
        {{ t('modemDetail.settings.aliasDescription') }}
      </p>
    </div>

    <div class="space-y-2">
      <Label for="modem-mss">{{ t('modemDetail.settings.mssLabel') }}</Label>
      <Input
        id="modem-mss"
        v-model="mss"
        type="number"
        min="64"
        max="254"
        step="1"
        :disabled="isInputDisabled"
        :placeholder="t('modemDetail.settings.mssPlaceholder')"
      />
      <p class="text-xs text-muted-foreground">
        {{ t('modemDetail.settings.mssDescription') }}
      </p>
    </div>

    <div class="space-y-3 rounded-lg border p-3">
      <div class="flex items-center justify-between gap-4">
        <div class="space-y-1">
          <Label for="modem-auto-unlock">{{ t('modemDetail.settings.autoUnlockLabel') }}</Label>
          <p class="text-xs text-muted-foreground">
            {{ t('modemDetail.settings.autoUnlockDescription') }}
          </p>
        </div>
        <Switch
          id="modem-auto-unlock"
          v-model="autoUnlock"
          :disabled="isInputDisabled"
        />
      </div>
      <div v-if="autoUnlock" class="space-y-2">
        <Label for="modem-sim-pin">{{ t('modemDetail.settings.simPinLabel') }}</Label>
        <Input
          id="modem-sim-pin"
          v-model="simPin"
          type="password"
          :disabled="isInputDisabled"
          :placeholder="t('modemDetail.settings.simPinPlaceholder')"
        />
        <p class="text-xs text-muted-foreground">
          {{
            simPinSaved
              ? t('modemDetail.settings.simPinSavedDescription')
              : t('modemDetail.settings.simPinDescription')
          }}
        </p>
      </div>
    </div>

    <div class="flex justify-end">
      <Button
        size="sm"
        type="button"
        class="w-full"
        :disabled="isActionDisabled"
        @click="emit('update')"
      >
        <span v-if="props.isUpdating" class="inline-flex items-center gap-2">
          <Spinner class="size-4" />
          {{ t('modemDetail.actions.update') }}
        </span>
        <span v-else>{{ t('modemDetail.actions.update') }}</span>
      </Button>
    </div>
  </section>
</template>
