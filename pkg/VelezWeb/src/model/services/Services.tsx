import MatreshkaIcon from "@/assets/icons/matreshka/MatreshkaIcon";
import MakoshIcon from "@/assets/icons/makosh/MakoshIcon";

export type Services = {
    matreshka?: Matreshka
    velez?: Svarog
    makosh?: Makosh
}


export class Service {
    tittle: string
    icon: React.JSX.Element
    externalLink: string

    constructor(tittle: string, icon: React.JSX.Element, externalLink: string) {
        this.tittle = tittle
        this.icon = icon
        this.externalLink = externalLink
    }

}

export class Makosh extends Service {
    constructor(externalLink: string) {
        super("Makosh", <MakoshIcon/>, externalLink);
    }
}

export class Svarog extends Service {
    constructor(externalLink: string) {
        super("Svarog", <MatreshkaIcon/>, externalLink);
    }
}

export class Matreshka extends Service {
    constructor(externalLink: string) {
        super("Matreshka", <MatreshkaIcon/>, externalLink);
    }
}

