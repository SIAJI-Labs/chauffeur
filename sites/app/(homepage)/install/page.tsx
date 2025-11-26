import { permanentRedirect } from 'next/navigation';

export const dynamic = 'force-static';

export default function InstallPage() {
  permanentRedirect('https://raw.githubusercontent.com/SIAJI-Labs/chauffeur/refs/heads/release/install.sh');
}