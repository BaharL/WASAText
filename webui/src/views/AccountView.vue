<template>
  <div class="pt-3">
    <h1 class="h3 mb-3">Account</h1>

    <div class="row">
      <!-- User info + change username -->
      <div class="col-md-6 mb-4">
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

              <p
                v-if="usernameError"
                class="text-danger mt-2 small"
              >
                {{ usernameError }}
              </p>
              <p
                v-if="usernameSuccess"
                class="text-success mt-2 small"
              >
                Username updated successfully.
              </p>
            </form>
          </div>
        </div>
      </div>

      <!-- Stub per foto profilo -->
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
 * - Change username
 * - Upload profile photo
 * - Sync photoUrl with sidebar
 */

import { computed, ref } from 'vue'
import { setMyUserName, setMyPhoto, getContext } from '../services/api.js'

// ⚠️ backend ORIGIN (uguale a quello dell’API)
const BACKEND_ORIGIN = 'http://localhost:3000'

const username = ref(localStorage.getItem('username') || '')
const newUsername = ref(username.value)

const savingUsername = ref(false)
const usernameError = ref('')
const usernameSuccess = ref(false)

const selectedPhoto = ref(null)
const uploadingPhoto = ref(false)
const photoMessage = ref('')
const photoError = ref(false)

const initial = computed(() =>
  username.value ? username.value.charAt(0).toUpperCase() : '?'
)

async function onChangeUsername() {
  usernameError.value = ''
  usernameSuccess.value = false

  const trimmed = newUsername.value.trim()
  if (!trimmed) return

  try {
    savingUsername.value = true
    await setMyUserName(trimmed)
    username.value = trimmed
    localStorage.setItem('username', trimmed)

    // aggiorna sidebar
    window.dispatchEvent(new Event('profile-updated'))
    usernameSuccess.value = true
  } catch (e) {
    console.error(e)
    usernameError.value = e.message || 'Failed to change username.'
  } finally {
    savingUsername.value = false
  }
}

function onPhotoSelected(event) {
  selectedPhoto.value = event.target.files?.[0] || null
  photoMessage.value = ''
  photoError.value = false
}

async function onUploadPhoto() {
  if (!selectedPhoto.value) return

  uploadingPhoto.value = true
  photoMessage.value = ''
  photoError.value = false

  try {
    // upload
    await setMyPhoto(selectedPhoto.value)

    // rileggo context
    const ctx = await getContext()

    if (ctx.photoUrl) {
      // 🔥 FIX FONDAMENTALE: URL ASSOLUTO
      const absoluteUrl = ctx.photoUrl.startsWith('http')
        ? ctx.photoUrl
        : BACKEND_ORIGIN + ctx.photoUrl

      localStorage.setItem('photoUrl', absoluteUrl)
    } else {
      localStorage.removeItem('photoUrl')
    }

    // aggiorna sidebar
    window.dispatchEvent(new Event('profile-updated'))

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
