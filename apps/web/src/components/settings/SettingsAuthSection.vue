<script setup lang="ts">
import { useI18n } from 'vue-i18n'

import { computed, ref, watch } from 'vue'
import SettingsChannelPicker from '@/components/settings/SettingsChannelPicker.vue'
import SettingsField from '@/components/settings/SettingsField.vue'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import type {
  SettingsAuth,
  SettingsChannelSchema,
  SettingsField as SettingsFieldSchema,
} from '@/types/settings'

const props = defineProps<{
  auth: SettingsAuth | null
  configuredChannels: SettingsChannelSchema[]
  fields: SettingsFieldSchema[]
  disabled?: boolean
}>()

const emit = defineEmits<{
  'update-field': [key: string, value: unknown]
  'update:valid': [valid: boolean]
}>()

const { t, te } = useI18n()
const otpChannels = computed(() => props.auth?.otpChannels ?? [])
const confirmPassword = ref('')
const passwordFieldValue = computed(() => fieldValue('password') as string | undefined)
const passwordMismatch = computed(() => {
  if (props.auth?.mode !== 'password') return false
  const confirm = confirmPassword.value
  return !confirm || confirm !== (passwordFieldValue.value ?? '')
})
watch(passwordMismatch, (mismatch) => emit('update:valid', !mismatch), { immediate: true })

const schemaText = (value: string | undefined) => {
  return value && te(value) ? t(value) : (value ?? '')
}

const fieldID = (key: string) => `settings-auth-${key}`

const fieldValue = (key: string) => {
  if (!props.auth) return undefined
  return key in props.auth ? props.auth[key as keyof SettingsAuth] : undefined
}

const visibleFields = computed(() =>
  props.fields.filter((field) => {
    if (field.key === 'mode') return true
    if (props.auth?.mode === 'cloudflare_access') return field.key.startsWith('access')
    if (props.auth?.mode === 'otp') return ['sessionMaxDays', 'otpChannels'].includes(field.key)
    if (props.auth?.mode === 'oidc') return field.key.startsWith('oidc')
    if (props.auth?.mode === 'password') return ['sessionMaxDays', 'password'].includes(field.key)
    return false
  }),
)
</script>

<template>
  <Card class="gap-4 border-0 py-4 shadow-sm">
    <CardContent class="space-y-4 px-4">
      <div
        v-for="field in visibleFields"
        :key="field.key"
        class="space-y-2"
      >
        <div
          v-if="field.control === 'channelList'"
          class="space-y-3"
        >
          <Label>{{ schemaText(field.label) }}</Label>
          <p
            v-if="field.description"
            class="text-xs text-muted-foreground"
          >
            {{ schemaText(field.description) }}
          </p>
          <SettingsChannelPicker
            :id="fieldID(field.key)"
            :model-value="otpChannels"
            :channels="configuredChannels"
            :disabled="disabled"
            @update:model-value="emit('update-field', 'otpChannels', $event)"
          />
        </div>

        <SettingsField
          v-else-if="field.key !== 'password'"
          :id="fieldID(field.key)"
          :field="field"
          :model-value="fieldValue(field.key)"
          :disabled="disabled"
          @update:model-value="emit('update-field', field.key, $event)"
        />

        <div
          v-else
          class="space-y-2"
        >
          <SettingsField
            :id="fieldID(field.key)"
            :field="field"
            :model-value="passwordFieldValue"
            :disabled="disabled"
            @update:model-value="emit('update-field', field.key, $event)"
          />
          <div class="space-y-2">
            <Label>{{ schemaText('settings.schema.auth.passwordConfirm') }}</Label>
            <Input
              id="settings-auth-password-confirm"
              type="password"
              :model-value="confirmPassword"
              :disabled="disabled"
              placeholder="••••••••"
              @update:model-value="confirmPassword = String($event)"
            />
            <p
              v-if="passwordMismatch"
              class="text-xs text-destructive"
            >
              {{ schemaText('settings.schema.auth.passwordMismatch') }}
            </p>
          </div>
        </div>
      </div>
    </CardContent>
  </Card>
</template>
