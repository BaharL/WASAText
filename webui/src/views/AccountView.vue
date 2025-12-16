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
              @change="onPhotoSelected"
              :disabled="uploadingPhoto"
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
 * ---------------
 * Page where the user can:
 * - View current username
 * - Change username (calls PUT /users/username)
 * - (Stub) select and upload a profile photo (PUT /users/photo → 501)
 */

import { computed, ref } from 'vue'
import { setMyUserName, setMyPhoto } from '../services/api.js'

const username = ref(localStorage.getItem('username') || '')
const newUsername = ref(username.value)

const savingUsername = ref(false)
const usernameError = ref('')
const usernameSuccess = ref(false)

const selectedPhoto = ref(null)
const uploadingPhoto = ref(false)
const photoMessage = ref('')
const photoError = ref(false)

const initial = computed(() => {
  const u = username.value
  return u ? u.charAt(0).toUpperCase() : '?'
})

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
    usernameSuccess.value = true
  } catch (e) {
    console.error(e)
    usernameError.value = e.message || 'Failed to change username.'
  } finally {
    savingUsername.value = false
  }
}

function onPhotoSelected(event) {
  const files = event.target.files
  selectedPhoto.value = files && files[0] ? files[0] : null
  photoMessage.value = ''
  photoError.value = false
}

async function onUploadPhoto() {
  if (!selectedPhoto.value) return

  photoMessage.value = ''
  photoError.value = false
  uploadingPhoto.value = true

  try {
    await setMyPhoto(selectedPhoto.value)
    photoMessage.value = 'Photo uploaded successfully.'
    photoError.value = false

    // ✅ 1) rileggo il context per ottenere la nuova photoUrl
    const res = await fetch('/v1/context', {
      headers: {
        Authorization: `Bearer ${localStorage.getItem('token')}`,
      },
    })
    if (res.ok) {
      const ctx = await res.json()
      if (ctx.photoUrl) {
        localStorage.setItem('photoUrl', ctx.photoUrl)
      } else {
        localStorage.removeItem('photoUrl')
      }
    }

    // ✅ 2) dico alla sidebar di aggiornarsi
    window.dispatchEvent(new Event('profile-updated'))

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
