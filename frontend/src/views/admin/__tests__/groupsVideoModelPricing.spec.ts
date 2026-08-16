import { describe, expect, it } from 'vitest'

import { groupPricingFromAPI, groupPricingToAPI } from '../GroupsView.vue'

describe('group video model pricing', () => {
  it('preserves video resolution tiers in the persisted payload', () => {
    const payload = groupPricingToAPI([{
      models: ['kling-v3-omni'],
      billing_mode: 'video',
      input_price: null,
      output_price: null,
      cache_write_price: null,
      cache_read_price: null,
      image_input_price: null,
      image_output_price: null,
      per_request_price: null,
      intervals: ['480p', '720p', '1080p'].map((tier_label, sort_order) => ({
        min_tokens: 0,
        max_tokens: null,
        tier_label,
        input_price: null,
        output_price: null,
        cache_write_price: null,
        cache_read_price: null,
        per_request_price: ['0.05', '0.06', '0.07'][sort_order],
        sort_order,
      })),
    }], 'openai')

    expect(payload[0].billing_mode).toBe('video')
    expect(payload[0].intervals.map(interval => interval.tier_label)).toEqual(['480p', '720p', '1080p'])
    expect(payload[0].intervals.map(interval => interval.per_request_price)).toEqual([0.05, 0.06, 0.07])
  })

  it('keeps image and video entries compatible when loading an edit form', () => {
    const entries = groupPricingFromAPI([
      {
        platform: 'openai', models: ['gpt-image-2'], billing_mode: 'image',
        input_price: null, output_price: null, cache_write_price: null, cache_read_price: null,
        image_input_price: null, image_output_price: null, per_request_price: 0.2, intervals: [],
      },
      {
        platform: 'openai', models: ['veo-3'], billing_mode: 'video',
        input_price: null, output_price: null, cache_write_price: null, cache_read_price: null,
        image_input_price: null, image_output_price: null, per_request_price: null,
        intervals: [{
          min_tokens: 0, max_tokens: null, tier_label: '720p',
          input_price: null, output_price: null, cache_write_price: null, cache_read_price: null,
          per_request_price: 0.08, sort_order: 0,
        }],
      },
    ])

    expect(entries.map(entry => entry.billing_mode)).toEqual(['image', 'video'])
    expect(entries[1].intervals[0].tier_label).toBe('720p')
  })
})
