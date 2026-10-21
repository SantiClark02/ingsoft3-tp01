// TP9: Sentry en el front (errores, demoras y uso que ve el navegador).
//
// El DSN de un front NO es un secreto: sólo sirve para MANDAR datos a Sentry, no para
// leerlos, y viaja igual dentro del JavaScript que baja cualquier navegador. Por eso va
// escrito acá (a diferencia del DSN del back, que llega por el entorno).
import * as Sentry from '@sentry/react';

const DSN_FRONT = 'https://33b5ae383c1201fbf00c12fc4dbc405e@o4512296739340288.ingest.us.sentry.io/4512296798388224';

// En el plan C los tres entornos son "localhost": los distingue el puerto.
const ENTORNOS = { '3001': 'qa', '3002': 'produccion', '3003': 'preprod' };

export function iniciarSentry() {
  if (!DSN_FRONT.startsWith('https://')) return; // sin DSN: apagado, la app anda igual
  Sentry.init({
    dsn: DSN_FRONT,
    environment: ENTORNOS[window.location.port] ?? 'local',
    integrations: [Sentry.browserTracingIntegration()],
    tracesSampleRate: 1.0, // medir todas las cargas y pedidos (app chica: entra en el plan gratis)
  });
}
