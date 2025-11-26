import { redirect } from 'next/navigation';

export const dynamic = 'force-static';

export default function InstallPage() {
  redirect('https://raw.githubusercontent.com/SIAJI-Labs/chauffeur/refs/heads/release/install.sh');
}