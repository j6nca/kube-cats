const config = {
    api_url: process.env.API_URL || 'http://localhost:8001',
    // api_url: process.env.API_URL || 'https://kube-cats-api.j6n.dev',
    environment: process.env.NODE_ENV || 'development',
    max_pods: process.env.MAX_PODS_PER_NODE || 200,
    show_labels: process.env.SHOW_LABELS || true,
    show_pod_count: process.env.SHOW_POD_COUNT || true,
    show_tooltips: process.env.SHOW_TOOLTIPS || false,
};

export default config;
