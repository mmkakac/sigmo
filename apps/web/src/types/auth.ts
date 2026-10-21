/** AuthMethods describes the backend's current trust boundary. */
export interface AuthMethods {
  version: number
  mode: 'none' | 'otp' | 'oidc' | 'cloudflare_access' | 'password'
  methods: Array<'otp' | 'oidc' | 'password'>
}

/** OTPChallenge binds a verification attempt to one delivery. */
export interface OTPChallenge {
  id: string
  expiresAt: string
}
