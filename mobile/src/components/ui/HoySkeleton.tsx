import React, { useEffect, useRef } from 'react';
import { View, Animated, Easing } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import BrandWordmark from './BrandWordmark';

// Skeleton en vez de spinner para la carga de contenido. Con el MISMO tiempo real
// de carga los estudios reportan ~30% de mejora en velocidad percibida: un
// spinner dirige la atencion a la espera y cada segundo se ve idéntico al
// anterior, mientras el skeleton la dirige al contenido que va a aparecer. Por
// eso replica el layout real de Hoy en vez de ser un bloque genérico.
//
// El chrome de las tarjetas si se conoce de antemano, asi que se dibuja de
// verdad; solo pulsa lo que todavia no sabemos.

const HAIRLINE = '#E7DECD';

function usePulse() {
  const value = useRef(new Animated.Value(0.45)).current;

  useEffect(() => {
    const loop = Animated.loop(
      Animated.sequence([
        Animated.timing(value, {
          toValue: 1,
          duration: 800,
          easing: Easing.inOut(Easing.ease),
          useNativeDriver: true,
        }),
        Animated.timing(value, {
          toValue: 0.45,
          duration: 800,
          easing: Easing.inOut(Easing.ease),
          useNativeDriver: true,
        }),
      ])
    );
    loop.start();
    return () => loop.stop();
  }, [value]);

  return value;
}

type BarProps = {
  opacity: Animated.Value;
  width: number | string;
  height: number;
  radius?: number;
  mt?: number;
};

// Estilos explicitos en vez de className: Animated.View no pasa por el
// transform de NativeWind de forma fiable.
function Bar({ opacity, width, height, radius = 6, mt = 0 }: BarProps) {
  return (
    <Animated.View
      style={{
        opacity,
        width: width as any,
        height,
        borderRadius: radius,
        backgroundColor: HAIRLINE,
        marginTop: mt,
      }}
    />
  );
}

export default function HoySkeleton() {
  const pulse = usePulse();

  return (
    <SafeAreaView className="flex-1 bg-cream" edges={['top']}>
      {/* Header: la marca ya se conoce, se dibuja de verdad. */}
      <View className="px-6 py-3 flex-row items-center justify-between border-b border-hairline/30 bg-cream">
        <BrandWordmark size="sm" />
        <Bar opacity={pulse} width={150} height={32} radius={16} />
        <Bar opacity={pulse} width={40} height={40} radius={20} />
      </View>

      <View className="px-6 pt-4">
        {/* El squad hoy */}
        <View className="rounded-3xl bg-paper border border-hairline p-5 mb-4">
          <Bar opacity={pulse} width={110} height={10} />
          <View className="flex-row gap-3 mt-4">
            <Bar opacity={pulse} width={48} height={48} radius={24} />
            <Bar opacity={pulse} width={48} height={48} radius={24} />
          </View>
        </View>

        {/* Countdown */}
        <View className="rounded-3xl bg-paper border border-hairline p-5 mb-4">
          <Bar opacity={pulse} width={130} height={10} />
          <Bar opacity={pulse} width={230} height={44} radius={8} mt={14} />
          <Bar opacity={pulse} width="100%" height={8} radius={4} mt={22} />
        </View>

        {/* Tu turno */}
        <View className="rounded-3xl bg-paper border border-hairline p-5 mb-4">
          <Bar opacity={pulse} width={80} height={10} />
          <Bar opacity={pulse} width={220} height={26} radius={6} mt={14} />
          <Bar opacity={pulse} width={170} height={14} radius={6} mt={16} />
          <Bar opacity={pulse} width={140} height={14} radius={6} mt={10} />
          <Bar opacity={pulse} width="100%" height={52} radius={26} mt={18} />
        </View>

        {/* Contadores */}
        <View className="rounded-3xl bg-paper border border-hairline flex-row overflow-hidden">
          {[0, 1, 2].map((i) => (
            <React.Fragment key={i}>
              {i > 0 ? <View className="w-[1px] bg-hairline" /> : null}
              <View className="flex-1 py-5 items-center">
                <Bar opacity={pulse} width={26} height={24} radius={6} />
                <Bar opacity={pulse} width={54} height={9} mt={8} />
              </View>
            </React.Fragment>
          ))}
        </View>
      </View>
    </SafeAreaView>
  );
}
