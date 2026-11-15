export function Isotipo({ tamano = 30 }: { tamano?: number }) {
  return (
    <svg width={tamano} height={tamano} viewBox="0 0 32 32" aria-hidden="true">
      <rect width="32" height="32" rx="8" fill="#1d4259" />
      <path d="M17 7v12a4 4 0 1 1-4-4" fill="none" stroke="#fff" strokeWidth="2.6" strokeLinecap="round" />
      <circle cx="17" cy="7" r="1.8" fill="#49c5a0" />
    </svg>
  );
}
