export const BASE_URL = "<URL>"

export const SITE_TITLE = "<TITLE>"
export const SITE_DESCRIPTION = "<DESCRIPTION>"

/*
 * Everything loads through this one GTM container. GA, pixels and any other
 * tags are configured inside GTM, not hardcoded here.
 */
export const GTM_ID: string = "<GTM-ID>"

/*
 * Optional page context, pushed to the dataLayer before the GTM loader runs so
 * the values are there when tags fire (it lands as dataLayer[0]). Leave it
 * empty to skip the push, e.g.:
 *
 *   { page_type: "landing", content_language: "en", page_id: "<PAGE-ID>" }
 */
export const GTM_PAGE_CONTEXT: Record<string, string> = {}
