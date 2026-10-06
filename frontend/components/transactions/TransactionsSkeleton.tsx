export default function TransactionsSkeleton() {
  return (
    <div className="flex flex-col gap-3">
      {[1, 2, 3].map((item) => (
        <div
          key={item}
          className="flex animate-pulse items-center justify-between rounded-2xl bg-white p-4"
        >
          <div className="flex items-center gap-3">
            <div className="h-11 w-11 rounded-full bg-[#dce9ff]" />

            <div className="flex flex-col gap-2">
              <div className="h-4 w-32 rounded bg-[#dce9ff]" />
              <div className="h-3 w-20 rounded bg-[#dce9ff]" />
            </div>
          </div>

          <div className="h-5 w-24 rounded bg-[#dce9ff]" />
        </div>
      ))}
    </div>
  );
}
