import { DownloadIcon, LinkIcon, Loader2, SearchIcon } from "lucide-react";
import { useEffect, useMemo, useRef, useState } from "react";
import { useFetch } from "./hooks/useFetch";
import { API_BASE_URL, fetchAPI } from "./api";
import slugify from "slugify";

const App = () => {
  const formRef = useRef<HTMLFormElement | null>(null);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [video, setVideo] = useState<any>(null);
  const [isLoading, setLoading] = useState(false);
  const tasks = useFetch("tasks", () => fetchAPI("/tasks/"));

  const onFindVideo = async () => {
    if (!formRef.current || isLoading) {
      return;
    }
    setErrors({});
    setVideo(null);
    setLoading(true);

    const formData = new FormData(formRef.current);
    const url = (formData.get("url") as string).trim();

    if (!url?.length || !url.startsWith("http")) {
      setErrors({ url: "URL is invalid" });
      setLoading(false);
      return;
    }

    try {
      const data = await fetchAPI("/info?url=" + encodeURI(url));
      setVideo(data);
    } catch (err) {
      setErrors({ url: (err as Error)?.message || "Unknown error" });
    }

    setLoading(false);
  };

  const onSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (isLoading) {
      return;
    }

    setErrors({});

    if (!video?.url) {
      setErrors({ url: "URL is invalid" });
      return;
    }

    const formData = new FormData(e.target as HTMLFormElement);
    const artist = (formData.get("artist") as string).trim();
    const slug = (formData.get("slug") as string).trim();
    const album = (formData.get("album") as string).trim();
    const title = (formData.get("title") as string).trim();

    const data = {
      url: video.url,
      thumbnail: video.thumbnail,
      title,
      slug,
      artist,
      album,
    };

    try {
      await fetchAPI("/tasks/", {
        method: "POST",
        body: JSON.stringify(data),
        headers: {
          "Content-Type": "application/json",
        },
      });

      // reset form
      formRef.current?.reset();
      setVideo(null);
      tasks.refetch();
    } catch (err) {
      setErrors({ result: (err as Error)?.message || "Unknown error" });
    }
  };

  useEffect(() => {
    const hasPendingTask = tasks.data?.find((task: any) => task.is_pending);
    if (hasPendingTask) {
      const timeout = setTimeout(() => tasks.refetch(), 1000);
      return () => clearTimeout(timeout);
    }
  }, [tasks.data]);

  return (
    <div className="min-h-screen flex flex-col items-center justify-center p-4 bg-base-300">
      <div className="card bg-base-100 w-full max-w-2xl">
        <div className="card-body p-4 md:p-8">
          <p className="card-title font-normal text-2xl self-center text-center">
            YouTube To MP3
          </p>
          <p className="self-center text-center">
            Download and convert YouTube videos to MP3
          </p>

          <form ref={formRef} onSubmit={onSubmit}>
            <label className="input input-bordered flex items-center gap-2 md:gap-3 pl-2 md:pl-4 pr-0 mt-4 overflow-hidden">
              <LinkIcon className="shrink-0" size={20} />
              <input
                className="grow"
                name="url"
                placeholder="Enter Video URL"
                required
                onKeyDown={(e) => {
                  if (e.key === "Enter") {
                    onFindVideo();
                    e.preventDefault();
                  }
                }}
              />
              <button
                type="button"
                className="btn btn-primary btn-square shrink-0"
                onClick={onFindVideo}
                disabled={isLoading}
              >
                {isLoading ? (
                  <Loader2 className="animate-spin" />
                ) : (
                  <SearchIcon />
                )}
              </button>
            </label>
            {errors.url && (
              <p className="text-error text-sm mt-1">{errors.url}</p>
            )}

            {video != null && (
              <div className="mt-4 md:mt-8">
                <img
                  src={video.thumbnail}
                  alt="thumbnail"
                  className="w-full md:max-w-[50%] rounded-box mx-auto"
                />

                <div className="mt-4 grid md:grid-cols-2 gap-2 md:gap-4">
                  <label className="form-control w-full max-w-xs">
                    <div className="label pt-0 pb-0.5">
                      <span className="label-text">Title</span>
                    </div>
                    <input
                      type="text"
                      name="title"
                      placeholder="Enter Title"
                      className="input input-bordered w-full"
                      defaultValue={video.title}
                      required
                      onChange={(e) => {
                        const slugEl = formRef.current?.querySelector(
                          'input[name="slug"]'
                        ) as HTMLInputElement | null;
                        if (slugEl) {
                          slugEl.value = slugify(e.target.value).toLowerCase();
                        }
                      }}
                    />
                  </label>

                  <label className="form-control w-full max-w-xs">
                    <div className="label pt-0 pb-0.5">
                      <span className="label-text">Slug</span>
                    </div>
                    <input
                      type="text"
                      name="slug"
                      placeholder="Enter Slug"
                      className="input input-bordered w-full"
                      defaultValue={video.slug}
                      required
                    />
                  </label>

                  <label className="form-control w-full max-w-xs">
                    <div className="label pt-0 pb-0.5">
                      <span className="label-text">Artist</span>
                    </div>
                    <input
                      type="text"
                      name="artist"
                      placeholder="Enter Artist"
                      className="input input-bordered w-full"
                      defaultValue={video.artist}
                    />
                  </label>

                  <label className="form-control w-full max-w-xs">
                    <div className="label pt-0 pb-0.5">
                      <span className="label-text">Album</span>
                    </div>
                    <input
                      type="text"
                      name="album"
                      placeholder="Enter Album"
                      className="input input-bordered w-full"
                      defaultValue={video.album}
                    />
                  </label>

                  <button
                    type="submit"
                    className="btn btn-primary mt-4 w-full max-w-xs"
                    disabled={isLoading}
                  >
                    {isLoading ? (
                      <Loader2 className="animate-spin" />
                    ) : (
                      <DownloadIcon />
                    )}
                    Download
                  </button>
                </div>
              </div>
            )}
          </form>
        </div>

        <TaskList data={tasks.data} />
      </div>
    </div>
  );
};

type TaskListProps = {
  data?: any[];
};

const TaskList = ({ data }: TaskListProps) => {
  const items = useMemo(() => {
    return data?.reverse() ?? [];
  }, [data]);

  if (!items?.length) {
    return null;
  }

  return (
    <div className="overflow-x-auto p-4 pt-0">
      <table className="table w-full">
        <thead>
          <tr>
            <th>Title</th>
            <th>Artist</th>
            <th>Album</th>
            <th>Status</th>
          </tr>
        </thead>
        <tbody>
          {items.map((task, idx) => {
            let downloadUrl = "";

            if (!task.is_pending && task.result) {
              const filename = task.result.split("/").pop();
              downloadUrl = API_BASE_URL + "/get/" + filename;
            }

            return (
              <tr key={task.slug + idx}>
                <td>
                  <div className="flex flex-row items-center gap-2">
                    <img
                      src={task.thumbnail}
                      alt="thumbnail"
                      className="w-8 h-8 object-cover"
                    />
                    {downloadUrl ? (
                      <a
                        href={downloadUrl}
                        target="_blank"
                        className="link truncate max-w-[200px]"
                        title={task.title}
                      >
                        {task.title}
                      </a>
                    ) : (
                      <span
                        className="truncate max-w-[200px]"
                        title={task.title}
                      >
                        {task.title}
                      </span>
                    )}
                  </div>
                </td>
                <td>
                  <span
                    className="inline-block truncate max-w-[160px]"
                    title={task.artist}
                  >
                    {task.artist || "-"}
                  </span>
                </td>
                <td>{task.album || "-"}</td>
                <td>
                  <div className="flex flex-row items-center gap-2">
                    <TaskStatus
                      isPending={task.is_pending}
                      error={task.error}
                    />
                    {downloadUrl ? (
                      <a
                        href={downloadUrl + "?dl=true"}
                        className="btn btn-ghost btn-square"
                      >
                        <DownloadIcon size={20} />
                      </a>
                    ) : null}
                  </div>
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
};

type TaskStatusProps = {
  isPending: boolean;
  error?: string | null;
};

const TaskStatus = ({ isPending, error }: TaskStatusProps) => {
  if (isPending) {
    return <Loader2 className="animate-spin" />;
  }

  return error ? (
    <p className="text-error text-sm" title={error}>
      Error!
    </p>
  ) : (
    <p className="text-success text-sm">Done</p>
  );
};

export default App;
