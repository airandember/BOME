INSERT INTO public_settings (key, value) VALUES ('youtube_channel_id', 'UCHp1EBgpKytZt_-j72EZ83Q') ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW();
