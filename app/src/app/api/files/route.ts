import * as k8s from "@kubernetes/client-node";

export const dynamic = "force-dynamic";

function getKubeConfig() {
  const kc = new k8s.KubeConfig();

  kc.addCluster({
    name: "my-cluster",
    server: "https://localhost:6443",
    skipTLSVerify: true,
  });

  kc.addContext({
    name: "my-context",
    cluster: "my-cluster",
    user: "my-user",
  });

  kc.setCurrentContext("my-context");
  return kc;
}

export async function GET() {
  const kc = getKubeConfig();
  const k8sApi = kc.makeApiClient(k8s.CustomObjectsApi);

  try {
    const filesRes = await k8sApi.listNamespacedCustomObject({
      group: "cdn.k8s.toms.place",
      version: "v1alpha1",
      namespace: "toms-place",
      plural: "files",
    });

    return Response.json(filesRes);
  } catch (error) {
    console.error("Failed to fetch files:", error);
    return Response.json({ error: String(error), items: [] }, { status: 500 });
  }
}
