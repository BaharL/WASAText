<template>
  <div class="pt-3">
    <h1 class="h3 mb-3">Account</h1>

    <div class="row">
      <div class="col-md-6 mb-4">
        <div class="card mb-3">
          <div class="card-body d-flex align-items-center">
            <div class="profile-avatar me-3">
              <img
                v-if="photoUrl"
                :src="toImgSrc(photoUrl)"
                alt="avatar"
                class="avatar-img"
              >
              <span v-else class="avatar-fallback">{{ initial }}</span>
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

              <p v-if="usernameError" class="text-danger mt-2 small">
                {{ usernameError }}
              </p>

              <p v-if="usernameSuccess" class="text-success mt-2 small">
                Username updated successfully.
              </p>
            </form>
          </div>
        </div>
      </div>

      <div class="col-md-6 mb-4">
        <div class="card">
          <div class="card-body">
            <div class="d-flex align-items-center justify-content-between mb-2">
              <h2 class="h5 mb-0">Profile photo</h2>

              <div class="profile-avatar profile-avatar-sm">
                <img
                  v-if="photoUrl"
                  :src="toImgSrc(photoUrl)"
                  alt="avatar"
                  class="avatar-img"
                >
                <span v-else class="avatar-fallback">{{ initial }}</span>
              </div>
            </div>

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
import { computed, ref } from 'vue'
import { setMyUserName, setMyPhoto, getContext } from '../services/api.js'

const username = ref(localStorage.getItem('username') || '')
const newUsername = ref(username.value)
const photoUrl = ref(localStorage.getItem('photoUrl') || '')

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

function notifyProfileUpdated() {
  window.dispatchEvent(new Event('profile-updated'))
}

// ✅ QUI È IL FIX: se in storage hai "/v1/..." allora va mostrato come "/api/v1/..."
function toImgSrc(url) {
  if (!url) return ''
  if (url.startsWith('/v1/')) return `/api${url}`
  return url
}

function normalizeUsernameError(err) {
  const raw = (err?.message || '').toLowerCase()
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

function extractPhotoUrlFromContext(ctx) {
  return (
    ctx?.photoUrl ||
    ctx?.photo_url ||
    ctx?.photo ||
    ctx?.avatarUrl ||
    ''
  )
}

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

    notifyProfileUpdated()
    usernameSuccess.value = true
  } catch (e) {
    console.error(e)
    usernameError.value = normalizeUsernameError(e)
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
    const uploadRes = await setMyPhoto(selectedPhoto.value)

    let finalUrl = ''
    try {
      const ctx = await getContext()
      finalUrl = extractPhotoUrlFromContext(ctx)
    } catch (_) {}

    if (!finalUrl) {
      finalUrl =
        uploadRes?.photoUrl ||
        uploadRes?.url ||
        uploadRes?.mediaUrl ||
        ''
    }

    if (finalUrl) {
      photoUrl.value = finalUrl
      localStorage.setItem('photoUrl', finalUrl)
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
/* identico al tuo */
.profile-avatar {
  width: 48px;
  height: 48px;
  border-radius: 999px;
  background: #0d6efd;
  color: #fff;
  overflow: hidden;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 1.3rem;
}
.profile-avatar-sm { width: 36px; height: 36px; font-size: 1rem; }
.avatar-img { width: 100%; height: 100%; object-fit: cover; }
.avatar-fallback { display: inline-flex; align-items: center; justify-content: center; }
</style>
