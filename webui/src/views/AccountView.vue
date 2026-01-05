<template>
  <div class="pt-3">
    <h1 class="h3 mb-3">Account</h1>

    <div class="row">
      <!-- =========================================================
           LEFT: Current user info + Change username form
           ========================================================= -->
      <div class="col-md-6 mb-4">
        <!-- Current user card -->
        <div class="card mb-3">
          <div class="card-body d-flex align-items-center">
            <div class="profile-avatar me-3">
              {{ initial }}
            </div>

            <div>
              <div class="text-muted small">Current username</div>
              <div class="fw-semibold">{{ username || 'unknown' }}</div>
            </div>
          </div>
        </div>

        <!-- Change username card -->
        <div class="card">
          <div class="card-body">
            <h2 class="h5 mb-3">Change username</h2>

            <form @submit.prevent="onChangeUsername">
              <div class="mb-3">
                <label for="username" class="form-label">New username</label>
                <input
                  id="username"
                  v-model.trim="newUsername"
                  type="text"
                  class="form-control"
                  placeholder="3–16 characters"
                  :disabled="savingUsername"
                >
              </div>

              <button
                type="submit"
                class="btn btn-primary"
                :disabled="savingUsername || !newUsername"
              >
                <span v-if="savingUsername">Saving…</span>
                <span v-else>Save username</span>
              </button>

              <!-- Error message (shown when API fails / conflict / validation) -->
              <p v-if="usernameError" class="text-danger mt-2 small">
                {{ usernameError }}
              </p>

              <!-- Success message -->
              <p v-if="usernameSuccess" class="text-success mt-2 small">
                Username updated successfully.
              </p>
            </form>
          </div>
        </div>
      </div>

      <!-- =========================================================
           RIGHT: Upload profile photo
           ========================================================= -->
      <div class="col-md-6 mb-4">
        <div class="card">
          <div class="card-body">
            <h2 class="h5 mb-3">Profile photo</h2>
            <p class="text-muted small mb-2">
              Upload a profile picture (JPG/PNG, max 10 MB).
            </p>

            <input
              type="file"
              accept="image/*"
              class="form-control mb-3"
              :disabled="uploadingPhoto"
              @change="onPhotoSelected"
            >

            <button
              type="button"
              class="btn btn-outline-secondary"
              :disabled="!selectedPhoto || uploadingPhoto"
              @click="onUploadPhoto"
            >
              <span v-if="uploadingPhoto">Uploading…</span>
              <span v-else>Upload photo</span>
            </button>

            <p
              v-if="photoMessage"
              class="mt-2 small"
              :class="photoError ? 'text-danger' : 'text-muted'"
            >
              {{ photoMessage }}
            </p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
/**
 * AccountView.vue
 *
 * Responsibilities:
 * - Display current username + avatar initial
 * - Allow user to change username
 * - Allow user to upload a profile photo
 *
 * Notes:
 * - We keep username/photoUrl in localStorage so Sidebar can render them quickly.
 * - After any update, we dispatch a "profile-updated" event so Sidebar can refresh.
 * - __API_URL__ is injected by Vite (in your project it's usually "/api").
 */

import { computed, ref } from 'vue'

// IMPORTANT:
// This file is under src/views/, so services are typically one level up: "../services/api.js"
import { setMyUserName, setMyPhoto, getContext } from '../services/api.js'

/* ------------------------------------------------------------------
 * STATE
 * ------------------------------------------------------------------ */

// Username state (loaded from localStorage at startup)
const username = ref(localStorage.getItem('username') || '')
const newUsername = ref(username.value)

// UI state for username change
const savingUsername = ref(false)
const usernameError = ref('')
const usernameSuccess = ref(false)

// UI state for photo upload
const selectedPhoto = ref(null)
const uploadingPhoto = ref(false)
const photoMessage = ref('')
const photoError = ref(false)

/* ------------------------------------------------------------------
 * COMPUTED
 * ------------------------------------------------------------------ */

// First letter shown in the blue circle (avatar placeholder)
const initial = computed(() =>
  username.value ? username.value.charAt(0).toUpperCase() : '?'
)

/* ------------------------------------------------------------------
 * HELPERS
 * ------------------------------------------------------------------ */

/**
 * Normalize backend errors into a clearer UI message.
 * We especially want a clean message when username is already taken.
 */
function normalizeUsernameError(err) {
  const raw = (err?.message || '').toLowerCase()

  // Detect typical "conflict / already exists" messages
  if (
    raw.includes('already') ||
    raw.includes('exists') ||
    raw.includes('duplicate') ||
    raw.includes('taken') ||
    raw.includes('conflict')
  ) {
    return 'This username is already taken. Please choose another one.'
  }

  return err?.message || 'Failed to change username.'
}

/**
 * Convert a returned photoUrl into a usable URL for the frontend.
 *
 * Why this exists:
 * - In your setup, __API_URL__ is usually "/api" (not a full origin).
 * - Backend could return:
 *    - absolute URL: "https://.../uploads/..."
 *    - API path: "/api/uploads/..."
 *    - plain path: "/uploads/..."
 *
 * We must avoid ending up with "/api/api/uploads/..." (double /api).
 */
function buildAbsolutePhotoUrl(photoUrl) {
  if (!photoUrl) return ''

  // Already absolute (http/https)
  if (photoUrl.startsWith('http')) return photoUrl

  // Already includes API prefix
  if (photoUrl.startsWith('/api/')) return photoUrl

  // Otherwise prefix with __API_URL__ (typically "/api")
  // Ensure we don't create missing/duplicate slashes
  const needsSlash = !photoUrl.startsWith('/')
  return `${__API_URL__}${needsSlash ? '/' : ''}${photoUrl}`
}

/**
 * Notify other UI pieces (e.g., Sidebar) that profile data changed.
 */
function notifyProfileUpdated() {
  window.dispatchEvent(new Event('profile-updated'))
}

/* ------------------------------------------------------------------
 * ACTIONS
 * ------------------------------------------------------------------ */

/**
 * Change username:
 * - Validate
 * - Call API
 * - Update local state + localStorage
 * - Notify Sidebar
 */
async function onChangeUsername() {
  usernameError.value = ''
  usernameSuccess.value = false

  const trimmed = newUsername.value.trim()
  if (!trimmed) return

  try {
    savingUsername.value = true

    // Backend call
    await setMyUserName(trimmed)

    // Update local state/storage
    username.value = trimmed
    localStorage.setItem('username', trimmed)

    notifyProfileUpdated()
    usernameSuccess.value = true
  } catch (e) {
    console.error(e)
    usernameError.value = normalizeUsernameError(e)
  } finally {
    savingUsername.value = false
  }
}

/**
 * Track selected image file.
 */
function onPhotoSelected(event) {
  selectedPhoto.value = event.target.files?.[0] || null
  photoMessage.value = ''
  photoError.value = false
}

/**
 * Upload photo:
 * - PUT /users/photo (multipart/form-data)
 * - Then reload /context to get the final photoUrl
 * - Store photoUrl in localStorage
 * - Notify Sidebar
 */
async function onUploadPhoto() {
  if (!selectedPhoto.value) return

  uploadingPhoto.value = true
  photoMessage.value = ''
  photoError.value = false

  try {
    // 1) Upload file
    await setMyPhoto(selectedPhoto.value)

    // 2) Reload context (server decides final URL/path)
    const ctx = await getContext()

    // 3) Save photoUrl locally (used by Sidebar)
    if (ctx.photoUrl) {
      const url = buildAbsolutePhotoUrl(ctx.photoUrl)
      localStorage.setItem('photoUrl', url)
    } else {
      localStorage.removeItem('photoUrl')
    }

    notifyProfileUpdated()
    photoMessage.value = 'Photo uploaded successfully.'
  } catch (e) {
    console.error(e)
    photoMessage.value = e.message || 'Upload failed.'
    photoError.value = true
  } finally {
    uploadingPhoto.value = false
  }
}
</script>

<style scoped>
.profile-avatar {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: #0d6efd;
  color: #fff;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 1.3rem;
}
</style>
