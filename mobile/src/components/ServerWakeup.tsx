import React, { useEffect, useState } from 'react';
import { View, Text, ActivityIndicator } from 'react-native';
import { supabase } from '../../lib/supabase';

// Espejo del ServerWakeup.vue de la web. Render free duerme un servicio a los 15
// min y su cold start es de ~13 s; mientras arranca, el edge de Render contesta
// 502 al instante, asi que el cliente reintenta hasta 39 s. Sin esto el movil
// mostraba un spinner pelon sobre fondo vacio todo ese rato, sin decir por que.
const BASE = process.env.EXPO_PUBLIC_API_URL || 'https://api-gateway-j3yi.onrender.com';

const WAKE_DEADLINE = 60_000; // ms — coincide con el "hasta un minuto" del copy
const POLL_INTERVAL = 2000; // ms — el 502 vuelve rapido, este es el ritmo real
const GRACE = 1500; // ms — un backend caliente nunca alcanza a mostrar el splash

// Cualquier respuesta que no sea 5xx prueba que el servicio ya esta procesando
// (un 401 tambien lo prueba, y no queremos insistir por un problema de auth).
async function isAwake(url: string, token?: string): Promise<boolean> {
  try {
    const res = await fetch(url, token ? { headers: { Authorization: `Bearer ${token}` } } : undefined);
    return res.ok || res.status < 500;
  } catch {
    return false; // red caida o todavia arrancando
  }
}

export default function ServerWakeup() {
  const [visible, setVisible] = useState(false);
  const [done, setDone] = useState(false);

  useEffect(() => {
    let cancelled = false;
    let awake = false;

    const graceTimer = setTimeout(() => {
      if (!awake && !cancelled) setVisible(true);
    }, GRACE);

    (async () => {
      const deadline = Date.now() + WAKE_DEADLINE;
      for (;;) {
        if (cancelled) return;

        const { data } = await supabase.auth.getSession();
        const token = data.session?.access_token;
        // Sin sesion solo podemos despertar al gateway: su middleware de Auth
        // rechaza /api/* antes de proxear, asi que una sonda anonima nunca llega
        // a habits. Con sesion despertamos groups y habits a la vez, que es lo
        // que necesita la pantalla de Hoy.
        const probes = token
          ? [isAwake(`${BASE}/api/groups`, token), isAwake(`${BASE}/api/habits`, token)]
          : [isAwake(`${BASE}/health`)];

        if ((await Promise.all(probes)).every(Boolean)) break;
        if (Date.now() >= deadline) break; // nunca dejar al usuario atrapado aqui
        await new Promise((r) => setTimeout(r, POLL_INTERVAL));
      }

      awake = true;
      if (cancelled) return;
      clearTimeout(graceTimer);
      setDone(true);
    })();

    return () => {
      cancelled = true;
      clearTimeout(graceTimer);
    };
  }, []);

  if (done || !visible) return null;

  return (
    <View
      style={{ position: 'absolute', top: 0, left: 0, right: 0, bottom: 0 }}
      className="bg-cream items-center justify-center px-10"
    >
      <ActivityIndicator size="large" color="#7CA39D" />
      <Text className="font-serif text-2xl font-semibold text-ink mt-8 mb-3 text-center">
        Despertando a tu squad…
      </Text>
      <Text className="text-sm text-ink-soft text-center leading-snug">
        El servidor estaba dormido y está arrancando. Esto puede tardar hasta un minuto.
      </Text>
    </View>
  );
}
