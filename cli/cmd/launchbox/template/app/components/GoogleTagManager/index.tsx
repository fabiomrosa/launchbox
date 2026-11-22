/*
 * next-script-for-ga wants @next/third-parties here, but that package pushes
 * `gtm.start` before the page context, loads through next/script's queue (the
 * snippet never appears in view-source) and has no <noscript> fallback. So the
 * snippets are written out verbatim, as Google documents them.
 */
/* eslint-disable @next/next/next-script-for-ga */
import { GTM_ID, GTM_PAGE_CONTEXT } from "@/config"

/*
 * Only rendered in production with a real container ID, to keep dev noise and
 * the "<GTM-ID>" placeholder out of GTM.
 */
const isEnabled =
  process.env.NODE_ENV === "production" &&
  GTM_ID.trim() !== "" &&
  GTM_ID !== "<GTM-ID>"

const hasPageContext = Object.keys(GTM_PAGE_CONTEXT).length > 0

/*
 * Goes first in <head>. Order matters: the page-context push runs before the
 * loader, so its values are on the dataLayer when tags fire.
 */
export function GoogleTagManagerHead() {
  if (!isEnabled) return null

  return (
    <>
      {hasPageContext && (
        <script
          dangerouslySetInnerHTML={{
            __html: `window.dataLayer = window.dataLayer || [];
window.dataLayer.push(${JSON.stringify(GTM_PAGE_CONTEXT)});`,
          }}
        />
      )}
      {/* Google Tag Manager */}
      <script
        dangerouslySetInnerHTML={{
          __html: `(function(w,d,s,l,i){w[l]=w[l]||[];w[l].push({'gtm.start':
new Date().getTime(),event:'gtm.js'});var f=d.getElementsByTagName(s)[0],
j=d.createElement(s),dl=l!='dataLayer'?'&l='+l:'';j.async=true;j.src=
'https://www.googletagmanager.com/gtm.js?id='+i+dl;f.parentNode.insertBefore(j,f);
})(window,document,'script','dataLayer','${GTM_ID}');`,
        }}
      />
      {/* End Google Tag Manager */}
    </>
  )
}

/* Goes first in <body>: the no-JavaScript fallback iframe. */
export function GoogleTagManagerNoScript() {
  if (!isEnabled) return null

  return (
    /* Google Tag Manager (noscript) */
    <noscript>
      <iframe
        src={`https://www.googletagmanager.com/ns.html?id=${GTM_ID}`}
        height="0"
        width="0"
        style={{ display: "none", visibility: "hidden" }}
      />
    </noscript>
  )
}
