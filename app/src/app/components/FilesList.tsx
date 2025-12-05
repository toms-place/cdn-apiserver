"use client";

import { useEffect, useState, useRef } from "react";
import type { components } from "../schema";

type FileType =
  components["schemas"]["place.toms.k8s.apiserver.pkg.apis.cdn.v1alpha1.File"];

interface WatchEvent {
  type: "ADDED" | "MODIFIED" | "DELETED" | "BOOKMARK" | "ERROR";
  object?: FileType;
  error?: string;
}

export function FilesList() {
  const [files, setFiles] = useState<Map<string, FileType>>(new Map());
  const [status, setStatus] = useState<"connecting" | "connected" | "error">(
    "connecting"
  );
  const [lastEvent, setLastEvent] = useState<string | null>(null);
  const initialFetchDone = useRef(false);

  useEffect(() => {
    // Fetch initial list only once
    if (!initialFetchDone.current) {
      initialFetchDone.current = true;
      fetch("/api/files")
        .then((res) => res.json())
        .then((data) => {
          const initialFiles = new Map<string, FileType>();
          for (const file of data.items || []) {
            if (file.metadata?.name) {
              initialFiles.set(file.metadata.name, file);
            }
          }
          setFiles(initialFiles);
        })
        .catch((error) => {
          console.error("Failed to fetch initial files:", error);
        });
    }

    // Set up watch stream using SSE (backed by @kubernetes/client-node on server)
    const eventSource = new EventSource("/api/files/watch");

    eventSource.onopen = () => {
      setStatus("connected");
    };

    eventSource.onmessage = (event) => {
      try {
        const watchEvent: WatchEvent = JSON.parse(event.data);
        setLastEvent(
          `${watchEvent.type}: ${
            watchEvent.object?.metadata?.name || "unknown"
          }`
        );

        if (watchEvent.type === "ERROR") {
          console.error("Watch error:", watchEvent.error);
          setStatus("error");
          return;
        }

        // BOOKMARK events are just for resourceVersion tracking, skip UI update
        if (watchEvent.type === "BOOKMARK") {
          return;
        }

        const fileName = watchEvent.object?.metadata?.name;
        if (!fileName) return;

        setFiles((prevFiles) => {
          const newFiles = new Map(prevFiles);

          switch (watchEvent.type) {
            case "ADDED":
            case "MODIFIED":
              newFiles.set(fileName, watchEvent.object!);
              break;
            case "DELETED":
              newFiles.delete(fileName);
              break;
          }

          return newFiles;
        });
      } catch (error) {
        console.error("Failed to parse watch event:", error);
      }
    };

    eventSource.onerror = () => {
      setStatus("error");
      // EventSource will automatically try to reconnect
    };

    return () => {
      eventSource.close();
    };
  }, []);

  const filesList = Array.from(files.values());

  return (
    <div className="w-full">
      <div className="mb-4 flex items-center gap-2">
        <span
          className={`inline-block h-3 w-3 rounded-full ${
            status === "connected"
              ? "bg-green-500"
              : status === "connecting"
              ? "bg-yellow-500 animate-pulse"
              : "bg-red-500"
          }`}
        />
        <span className="text-sm text-zinc-600 dark:text-zinc-400">
          {status === "connected"
            ? "Watching for changes..."
            : status === "connecting"
            ? "Connecting..."
            : "Connection error"}
        </span>
      </div>

      {lastEvent && (
        <div className="mb-4 text-xs text-zinc-500 dark:text-zinc-500">
          Last event: {lastEvent}
        </div>
      )}

      <h2 className="text-2xl font-semibold text-zinc-900 dark:text-white sm:text-3xl">
        Files in &quot;toms-place&quot; Namespace:
      </h2>

      {filesList.length === 0 ? (
        <p className="mt-4 text-zinc-600 dark:text-zinc-400">No files found.</p>
      ) : (
        <ul className="mt-4 space-y-2">
          {filesList.map((file) => (
            <li
              key={file.metadata?.name}
              className="rounded-lg border border-zinc-200 bg-zinc-50 p-3 dark:border-zinc-800 dark:bg-zinc-900"
            >
              <div className="font-medium text-zinc-900 dark:text-white">
                {file.metadata?.name}
              </div>
              {file.metadata?.creationTimestamp && (
                <div className="text-xs text-zinc-500 dark:text-zinc-500">
                  Created:{" "}
                  {new Date(file.metadata.creationTimestamp).toLocaleString()}
                </div>
              )}
              {file.metadata?.resourceVersion && (
                <div className="text-xs text-zinc-500 dark:text-zinc-500">
                  Version: {file.metadata.resourceVersion}
                </div>
              )}
            </li>
          ))}
        </ul>
      )}

      <div className="mt-4 text-sm text-zinc-500 dark:text-zinc-500">
        Total: {filesList.length} file(s)
      </div>
    </div>
  );
}
