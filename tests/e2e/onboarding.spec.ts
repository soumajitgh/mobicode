import { expect, test } from '@playwright/test';

test('first boot through authenticated browser access', async ({ page }) => {
  await page.goto('/');
  await expect(page).toHaveURL(/\/onboarding$/);

  await page.getByLabel('Email address').fill('owner@example.com');
  await page.getByLabel('Password', { exact: true }).fill('validpassword123');
  await page.getByLabel('Confirm password').fill('validpassword123');
  await page.getByRole('button', { name: 'Create account and continue' }).click();
  await expect(page).toHaveURL(/\/onboarding\?step=2$/);

  await page.getByRole('link', { name: 'Continue' }).click();
  await expect(page).toHaveURL(/\/onboarding\?step=3$/);
  await page.getByRole('button', { name: 'Finish setup' }).click();
  await expect(page).toHaveURL(/\/$/);
  await expect(page.getByRole('heading', { name: 'MobiCode' })).toBeVisible();

  await page.getByRole('button', { name: 'Log out' }).click();
  await expect(page).toHaveURL(/\/auth\/login$/);

  await page.goto('/');
  await expect(page).toHaveURL(/\/auth\/login\?next=/);
  await page.getByLabel('Email').fill('owner@example.com');
  await page.getByLabel('Password').fill('validpassword123');
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page).toHaveURL(/\/$/);
});
