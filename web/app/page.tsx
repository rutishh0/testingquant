import {
  Card,  CardContent,  CardDescription,  CardHeader,  CardTitle,} from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import Preprocess from "@/components/construction-preprocess";
import Payloads from "@/components/construction-payloads";
import Combine from "@/components/construction-combine";
import Submit from "@/components/construction-submit";

export default function Home() {
  return (
    <main className="flex min-h-screen flex-col items-center justify-between p-24">
      <div className="z-10 w-full max-w-5xl items-center justify-between font-mono text-sm lg:flex">
        <h1 className="text-4xl font-bold text-center">Coinbase Mesh Connector Playground</h1>
      </div>

      <div className="w-full max-w-5xl mt-16">
        <Tabs defaultValue="preprocess">
          <TabsList className="grid w-full grid-cols-4">
            <TabsTrigger value="preprocess">Preprocess</TabsTrigger>
            <TabsTrigger value="payloads">Payloads</TabsTrigger>
            <TabsTrigger value="combine">Combine</TabsTrigger>
            <TabsTrigger value="submit">Submit</TabsTrigger>
          </TabsList>
          <TabsContent value="preprocess">
            <Card>
              <CardHeader>
                <CardTitle>Construction Preprocess</CardTitle>
                <CardDescription>
                  Preprocess a list of operations to calculate balances changes.
                </CardDescription>
              </CardHeader>
              <CardContent>
                <Preprocess />
              </CardContent>
            </Card>
          </TabsContent>
          <TabsContent value="payloads">
            <Card>
              <CardHeader>
                <CardTitle>Construction Payloads</CardTitle>
                <CardDescription>
                  Generate an unsigned transaction and signing payloads.
                </CardDescription>
              </CardHeader>
              <CardContent>
                <Payloads />
              </CardContent>
            </Card>
          </TabsContent>
          <TabsContent value="combine">
            <Card>
              <CardHeader>
                <CardTitle>Construction Combine</CardTitle>
                <CardDescription>
                  Combine an unsigned transaction and signatures into a signed transaction.
                </CardDescription>
              </CardHeader>
              <CardContent>
                <Combine />
              </CardContent>
            </Card>
          </TabsContent>
          <TabsContent value="submit">
            <Card>
              <CardHeader>
                <CardTitle>Construction Submit</CardTitle>
                <CardDescription>Submit a signed transaction to the network.</CardDescription>
              </CardHeader>
              <CardContent>
                <Submit />
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </div>
    </main>
  );
}

