'use client';

import { Add } from '@material-symbols-svg/react/w400';

export default function AddTransactionButton() {
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
      <button
        type="button"
        aria-label="Add transaction"
        className="
          pointer-events-auto
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
    </div>
  );
}
