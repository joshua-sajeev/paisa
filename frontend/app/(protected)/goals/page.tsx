import Link from "next/link";
import { AdsClick } from "@material-symbols-svg/react/w400";

export default function GoalsPage() {
  return (
    <main className="mx-auto flex min-h-full w-full max-w-2xl flex-col justify-center p-5">
      <section className="rounded-3xl border border-slate-200 bg-white px-6 py-10 text-center shadow-sm sm:px-10">
        <div className="mx-auto flex h-14 w-14 items-center justify-center rounded-2xl bg-blue-50 text-blue-600">
          <AdsClick className="h-7 w-7" />
        </div>
        <p className="mt-5 text-xs font-bold uppercase tracking-[0.18em] text-blue-600">
          Coming soon
        </p>
        <h1 className="mt-2 text-2xl font-extrabold tracking-tight text-slate-900">
          Your goals are still goal-setting.
        </h1>
        <p className="mx-auto mt-2 max-w-md text-sm leading-6 text-slate-500">
          I am getting this space ready so you can plan what you’re saving for
          and watch your progress grow.
        </p>
        <Link
          href="/dashboard"
          className="mt-6 inline-flex items-center justify-center rounded-full bg-slate-900 px-5 py-2.5 text-sm font-bold text-white transition hover:bg-slate-800"
        >
          Back to dashboard
        </Link>
      </section>
    </main>
  );
}
