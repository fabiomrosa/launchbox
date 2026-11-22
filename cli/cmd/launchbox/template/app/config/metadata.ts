type ShareImageMetadata = {
  alt: string
  url: string
  width: number
  height: number
}

type OpenGraphMetadata = {
  type: string
  locale: string
  siteName: string
  url: string
  title: string
  description: string
  images: ShareImageMetadata[] | ShareImageMetadata
}

type MetadataOptions = {
  baseURL: string
  siteTitle: string
  siteDescription: string
  siteName?: string | null
  ogSiteTitle?: string | null
  ogSiteDescription?: string | null
  ogShareImage?: string | null
  ogShareImageAlt?: string | null
}

type Metadata = {
  metadataBase?: URL
  title: string
  description: string
  openGraph?: OpenGraphMetadata
}

export function getMetadata({
  baseURL,
  siteTitle,
  siteDescription,
  siteName = null,
  ogSiteTitle = null,
  ogSiteDescription = null,
  ogShareImage = null,
  ogShareImageAlt = null,
}: Readonly<MetadataOptions>): Metadata {
  const defaultOpenGraphImage = `${baseURL}/images/opengraph-image.jpg`

  return {
    // Skipped while BASE_URL is still the "<URL>" placeholder, which isn't a valid URL
    ...(URL.canParse(baseURL) && { metadataBase: new URL(baseURL) }),
    title: siteTitle,
    description: siteDescription,
    openGraph: {
      locale: "en_US",
      type: "website",
      siteName: siteName ?? siteTitle,
      url: baseURL,
      title: ogSiteTitle ?? siteTitle,
      description: ogSiteDescription ?? siteDescription,
      images: [
        {
          alt: ogShareImageAlt ?? ogSiteTitle ?? siteTitle,
          url: ogShareImage ?? defaultOpenGraphImage,
          width: 1200,
          height: 630,
        },
      ],
    },
  }
}
