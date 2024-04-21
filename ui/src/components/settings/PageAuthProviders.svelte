<script>
    import ApiClient from "@/utils/ApiClient";
    import { pageTitle } from "@/stores/app";
    import PageWrapper from "@/components/base/PageWrapper.svelte";
    import SettingsSidebar from "@/components/settings/SettingsSidebar.svelte";
    import AuthProviderCard from "@/components/settings/AuthProviderCard.svelte";
    import providersList from "@/providers.js";

    $pageTitle = "Auth providers";

    let isLoading = false;
    let formSettings = {};

    $: enabledProviders = providersList.filter((provider) => formSettings[provider.key]?.enabled);

    $: disabledProviders = providersList.filter((provider) => !formSettings[provider.key]?.enabled);

    loadSettings();

    async function loadSettings() {
        isLoading = true;

        try {
            const result = (await ApiClient.settings.getAll()) || {};
            alert(result);
            initSettings(result);
        } catch (err) {
            alert(err);
            ApiClient.error(err);
        }

        isLoading = false;
    }

    function initSettings(data) {
        data = data || {};
        formSettings = {};

        for (const provider of providersList) {
            formSettings[provider.key] = Object.assign({ enabled: false }, data[provider.key]);
        }
        alert(formSettings);
    }
</script>

<SettingsSidebar />

<PageWrapper>
    <header class="page-header">
        <nav class="breadcrumbs">
            <div class="breadcrumb-item">Settings</div>
            <div class="breadcrumb-item">{$pageTitle}</div>
        </nav>
    </header>

    <div class="wrapper">

    </div>
</PageWrapper>
