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
  const watch = new k8s.Watch(kc);

  const encoder = new TextEncoder();
  let abortController: AbortController | null = null;

  const stream = new ReadableStream({
    async start(controller) {
      try {
        abortController = await watch.watch(
          "/apis/cdn.k8s.toms.place/v1alpha1/namespaces/toms-place/files",
          {},
          (type, apiObj) => {
            const event = {
              type,
              object: apiObj,
            };
            controller.enqueue(
              encoder.encode(`data: ${JSON.stringify(event)}\n\n`)
            );
          },
          (err) => {
            if (err) {
              console.error("Watch error:", err);
              controller.enqueue(
                encoder.encode(
                  `data: ${JSON.stringify({
                    type: "ERROR",
                    error: String(err),
                  })}\n\n`
                )
              );
            }
            controller.close();
          }
        );
      } catch (err) {
        console.error("Failed to start watch:", err);
        controller.enqueue(
          encoder.encode(
            `data: ${JSON.stringify({ type: "ERROR", error: String(err) })}\n\n`
          )
        );
        controller.close();
      }
    },
    cancel() {
      if (abortController) {
        abortController.abort();
      }
    },
  });

  return new Response(stream, {
    headers: {
      "Content-Type": "text/event-stream",
      "Cache-Control": "no-cache",
      Connection: "keep-alive",
    },
  });
}
