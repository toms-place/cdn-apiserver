import { FilesList } from "./components/FilesList";

export default function Home() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-zinc-50 font-sans dark:bg-black">
      <main className="flex min-h-screen w-full max-w-3xl flex-col items-center justify-between py-32 px-16 bg-white dark:bg-black sm:items-start">
        <h1 className="text-4xl font-bold text-zinc-900 dark:text-white sm:text-5xl">
          Kubernetes Custom Resource Viewer
        </h1>
        <div className="mt-10 w-full">
          <FilesList />
        </div>
      </main>
    </div>
  );
}
