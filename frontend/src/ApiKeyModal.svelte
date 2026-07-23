<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte'
  import {
    HasDeepSeekAPIKey,
    SaveDeepSeekAPIKey,
    RemoveDeepSeekAPIKey,
    HasUnsplashAPIKey,
    SaveUnsplashAPIKey,
    RemoveUnsplashAPIKey
  } from '../wailsjs/go/main/CredentialStore'
  import { logError, logInfo } from './logger'

  const dispatch = createEventDispatcher()

  let deepSeekApiKey = ''
  let hasDeepSeekApiKey = false
  let isSavingDeepSeek = false
  let saveDeepSeekError = ''
  let deepSeekStatusMessage = ''

  let unsplashApiKey = ''
  let hasUnsplashApiKey = false
  let isSavingUnsplash = false
  let saveUnsplashError = ''
  let unsplashStatusMessage = ''

  async function checkApiKeys() {
    try {
      hasDeepSeekApiKey = await HasDeepSeekAPIKey()
      deepSeekStatusMessage = hasDeepSeekApiKey
        ? 'A DeepSeek API key is configured.'
        : 'No DeepSeek API key is configured.'
    } catch (error) {
      logError('Could not check for DeepSeek API key', error)
      deepSeekStatusMessage = 'Could not check for DeepSeek API key.'
    }

    try {
      hasUnsplashApiKey = await HasUnsplashAPIKey()
      unsplashStatusMessage = hasUnsplashApiKey
        ? 'An Unsplash API key is configured.'
        : 'No Unsplash API key is configured.'
    } catch (error) {
      logError('Could not check for Unsplash API key', error)
      unsplashStatusMessage = 'Could not check for Unsplash API key.'
    }
  }

  async function saveDeepSeek() {
    isSavingDeepSeek = true
    saveDeepSeekError = ''
    try {
      await SaveDeepSeekAPIKey(deepSeekApiKey)
      logInfo('DeepSeek API key saved')
      deepSeekApiKey = ''
      await checkApiKeys()
    } catch (error) {
      logError('Could not save DeepSeek API key', error)
      saveDeepSeekError = 'Could not save DeepSeek API key.'
    } finally {
      isSavingDeepSeek = false
    }
  }

  async function removeDeepSeek() {
    try {
      await RemoveDeepSeekAPIKey()
      logInfo('DeepSeek API key removed')
      await checkApiKeys()
    } catch (error) {
      logError('Could not remove DeepSeek API key', error)
      saveDeepSeekError = 'Could not remove DeepSeek API key.'
    }
  }

  async function saveUnsplash() {
    isSavingUnsplash = true
    saveUnsplashError = ''
    try {
      await SaveUnsplashAPIKey(unsplashApiKey)
      logInfo('Unsplash API key saved')
      unsplashApiKey = ''
      await checkApiKeys()
    } catch (error) {
      logError('Could not save Unsplash API key', error)
      saveUnsplashError = 'Could not save Unsplash API key.'
    } finally {
      isSavingUnsplash = false
    }
  }

  async function removeUnsplash() {
    try {
      await RemoveUnsplashAPIKey()
      logInfo('Unsplash API key removed')
      await checkApiKeys()
    } catch (error) {
      logError('Could not remove Unsplash API key', error)
      saveUnsplashError = 'Could not remove Unsplash API key.'
    }
  }

  function close() {
    dispatch('close')
  }

  onMount(() => {
    void checkApiKeys()
  })
</script>

<div class="modal-scrim" on:click={close}>
  <div class="modal-content" on:click|stopPropagation>
    <header class="modal-header">
      <h2>Configure API Keys</h2>
      <button class="close-button" type="button" on:click={close}>×</button>
    </header>
    <div class="modal-body">
      <section>
        <h3>DeepSeek</h3>
        <p>{deepSeekStatusMessage}</p>
        {#if hasDeepSeekApiKey}
          <button class="button-danger" on:click={removeDeepSeek}>Remove API Key</button>
        {:else}
          <form on:submit|preventDefault={saveDeepSeek}>
            <label for="deepSeekApiKey">API Key</label>
            <input id="deepSeekApiKey" type="password" bind:value={deepSeekApiKey} required />
            <button type="submit" disabled={isSavingDeepSeek}>
              {isSavingDeepSeek ? 'Saving...' : 'Save API Key'}
            </button>
            {#if saveDeepSeekError}
              <p class="error">{saveDeepSeekError}</p>
            {/if}
          </form>
        {/if}
      </section>
      <section>
        <h3>Unsplash</h3>
        <p>{unsplashStatusMessage}</p>
        {#if hasUnsplashApiKey}
          <button class="button-danger" on:click={removeUnsplash}>Remove API Key</button>
        {:else}
          <form on:submit|preventDefault={saveUnsplash}>
            <label for="unsplashApiKey">Access Key</label>
            <input id="unsplashApiKey" type="password" bind:value={unsplashApiKey} required />
            <button type="submit" disabled={isSavingUnsplash}>
              {isSavingUnsplash ? 'Saving...' : 'Save Access Key'}
            </button>
            {#if saveUnsplashError}
              <p class="error">{saveUnsplashError}</p>
            {/if}
          </form>
        {/if}
      </section>
    </div>
  </div>
</div>

<style>
  .modal-scrim {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-color: rgba(0, 0, 0, 0.5);
    display: flex;
    justify-content: center;
    align-items: center;
  }

  .modal-content {
    background-color: rgba(20, 20, 30, 0.92);
    backdrop-filter: blur(16px);
    -webkit-backdrop-filter: blur(16px);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 8px;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.35);
    width: 90%;
    max-width: 500px;
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem 1.5rem;
    border-bottom: 1px solid var(--color-border);
  }

  .modal-header h2 {
    margin: 0;
    font-size: 1.25rem;
  }

  .close-button {
    background: none;
    border: none;
    font-size: 1.5rem;
    cursor: pointer;
    color: var(--color-text-muted);
  }

  .modal-body {
    padding: 1.5rem;
    display: flex;
    flex-direction: column;
    gap: 2rem;
  }

  form {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .button-danger {
    background-color: var(--color-danger);
    color: white;
  }

  .error {
    color: var(--color-danger);
  }
</style>
