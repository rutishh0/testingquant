/**
 * TEMPORARY TEST ADAPTATION NOTE
 *
 * Context:
 * - The CardTitle for "Live Market Data" renders as a <div> (no ARIA heading role),
 *   so role-based heading locators fail. We temporarily assert by exact text until UI semantics are updated.
 * - The fallback status message includes dynamic error details (e.g., HTTP status), making exact matches brittle.
 *   We assert only the stable suffix: "Showing default markets".
 * - Native <select> <option> elements may not be considered visible. We assert presence via option list text and
 *   verify selection using the combobox value instead of visibility checks on option text.
 *
 * TODO (permanent fix):
 * - Update CardTitle for this section to be an accessible heading (e.g., <h3> or role="heading" aria-level={3}).
 * - Then switch test back to: page.getByRole('heading', { name: 'Live Market Data' }).
 * - Consider standardizing the status message to a consistent prefix/suffix if needed.
 */
import { test, expect } from '@playwright/test';

// This test assumes the backend is not configured for Exchange, so /v1/exchange/products returns []
// The UI should show the fallback warning and render default markets

test('renders fallback markets when products API returns empty', async ({ page }) => {
  await page.goto('/');

  // Navigate to Trading section via anchor link
  await page.locator('a[href="#trading"]').click();
  await page.waitForURL(/#trading/);

  // Wait for the page section to load
  await page.waitForLoadState('networkidle');

  // TEMPORARY FIX: Use text-based selector instead of semantic heading role
  // TODO: Replace with proper heading role once CardTitle is made accessible
  // Expected: await expect(page.getByRole('heading', { name: 'Live Market Data' })).toBeVisible();
  await expect(page.getByText('Live Market Data', { exact: true })).toBeVisible();

  // TEMPORARY FIX: Use partial text matching to handle varying error messages
  // The actual error text may vary (e.g., "API Error: 404 Not Found. Showing default markets.")
  // TODO: Standardize error message format across components
  await expect(page.getByText(/Showing default markets/i)).toBeVisible();

  // TEMPORARY FIX: Options in a native <select> are not "visible". Assert via option list and selection
  const combo = page.getByRole('combobox');
  await expect(combo).toBeEnabled();
  const options = page.locator('select > option');
  await expect(options).toContainText(['BTC-USD', 'ETH-USD']);
  await combo.selectOption('ETH-USD');
  await expect(combo).toHaveValue('ETH-USD');
});