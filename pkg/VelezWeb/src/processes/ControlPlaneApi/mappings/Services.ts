import {
    Matreshka as ApiMatreshka,
    Makosh as ApiMakosh,
    Svarog as ApiSvarog,
    ListServicesResponse as ApiServices
} from "@godverv/velez";

import {
    Matreshka,
    Makosh,
    Svarog,
    Services, Service,
} from "@/model/services/Services";

export function mapServices(services: ApiServices): Services {
    return {
        matreshka: mapMatreshka(services.matreshka),
        makosh: mapMakosh(services.makosh),
        svarog: mapSvarog(services.svarog),
    } as Services;
}

export function toServices(services: ApiServices): Service[] {
    const out: Service[] = []

    const m = mapMatreshka(services.matreshka)
    if (m) {
        out.push(m)
    }

    const s = mapMakosh(services.makosh)
    if (s) {
        out.push(s)
    }

    const sv = mapSvarog(services.svarog)
    if (sv) {
        out.push(sv)
    }

    return out
}

export function mapMatreshka(apiComponent?: ApiMatreshka): Matreshka | undefined {
    if (!apiComponent) {
        return
    }

    return new Matreshka("")
}

export function mapSvarog(apiComponent?: ApiMakosh): Makosh | undefined {
    if (!apiComponent) {
        return
    }

    return new Svarog("")
}

export function mapMakosh(apiComponent?: ApiSvarog): Svarog | undefined {
    if (!apiComponent) {
        return
    }

    return new Makosh("")
}
