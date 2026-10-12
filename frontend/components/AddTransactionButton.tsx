'use client';

import Link from 'next/link';
import { usePathname, useSearchParams } from 'next/navigation';
import { Add } from '@material-symbols-svg/react/w400';

export default function AddTransactionButton() {
  const pathname = usePathname();
  const searchParams = useSearchParams();

  if (pathname === '/transactions/add') {
    return null;
  }

  const dateParam = searchParams.get('date');
  const targetHref = dateParam ? `/transactions/add?date=${dateParam}` : `/transactions/add`;

  return (
    <div
      className="
        fixed
        left-1/2
        bottom-15
        z-[60]
        -translate-x-1/2
        pointer-events-none
      "
    >
      <Link href={targetHref} className="pointer-events-auto">
        <button
          type="button"
          aria-label="Add transaction"
          className="
            flex
            h-[60px]
            w-[60px]
            items-center
            justify-center
            rounded-2xl
            bg-[#2066E6]
            text-white
            shadow-[0_10px_25px_-5px_rgba(6,127,243,0.5),0_4px_10px_rgba(0,0,0,0.08)]
            transition-transform
            active:scale-90
          "
        >
          <Add size={32} color="currentColor" />
        </button>
      </Link>
    </div>
  );
}
