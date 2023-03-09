import React from 'react';
import { useQuery } from "react-query";

export interface BankLogoProps {
  plaidInstitutionId?: string;
}

export default function BankLogo(props: BankLogoProps): JSX.Element {
  const { data } = useQuery<{ logo: string }>(`/institutions/${ props.plaidInstitutionId }`, {
    enabled: !!props.plaidInstitutionId,
    staleTime: 60 * 60 * 1000, // 60 minutes
  });

  if (!data || !data?.logo) {
    return null;
  }

  return (
    <img
      className="max-h-8 col-span-4"
      src={ `data:image/png;base64,${ data.logo }` }
    />
  )
}
